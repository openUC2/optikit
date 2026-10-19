package gltf

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"slices"

	"github.com/pkg/errors"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
	"github.com/ungerik/go3d/float64/vec3"

	"github.com/openUC2/optikit/exp/designs"
	ffs "github.com/openUC2/optikit/exp/fs"
	"github.com/openUC2/optikit/exp/structures"
)

type Document struct {
	d    *gltf.Document
	root *gltf.Node

	// modelInstances contains the lists of node indices of the root nodes of all instances of each
	// primitive model added to the document. The key is a hash of each model (for
	// content-addressability).
	modelInstances map[string][][]int
	// indexMapping contains the index mappings of array elements for all primitive models added to
	// the document. The key is a hash of each model (for content-addressability).
	indexMappings map[string]indexMappings
}

// indexMappings maps from the indices of array elements within a model to the indices of those
// same array elements within the document to which that model has been added.
type indexMappings struct {
	Accessors map[int]int
	Materials map[int]int
	Meshes    map[int]int
	Nodes     map[int]int
}

// Document

func NewDocument() *Document {
	d := Document{
		d:              gltf.NewDocument(),
		root:           &gltf.Node{},
		modelInstances: make(map[string][][]int),
		indexMappings:  make(map[string]indexMappings),
	}
	d.d.Asset = gltf.Asset{
		Generator: "optikit", // TODO: add a version number!
		Version:   "2.0",
	}
	d.d.Nodes = []*gltf.Node{d.root}
	d.d.Scenes = []*gltf.Scene{
		{
			Name:  "design",
			Nodes: []int{0},
		},
	}
	return &d
}

func Load(contents []byte) (doc *Document, err error) {
	doc = &Document{
		d: gltf.NewDocument(),
	}
	if err = gltf.NewDecoder(bytes.NewReader(contents)).Decode(doc.d); err != nil {
		return nil, err
	}
	return doc, nil
}

func (d *Document) Document() *gltf.Document {
	return d.d
}

func (d *Document) Assemble(
	ctx context.Context,
	design *designs.FSDesign, assembly designs.AssmID, gridSpacings designs.ContinuousXYZ[float64],
	asText bool,
) (encoded []byte, err error) {
	d.root.Name = string(assembly)
	if design, err = design.Canonicalized(ctx, gridSpacings); err != nil {
		return nil, errors.Wrap(err, "couldn't canonicalize design")
	}
	if err = d.addComponents(design, assembly, d.root, gridSpacings); err != nil {
		return nil, err
	}

	if asText {
		return d.Encode(true)
	}

	return d.Encode(false)
}

func (d *Document) addComponents(
	design *designs.FSDesign, assembly designs.AssmID,
	root *gltf.Node,
	gridSpacings designs.ContinuousXYZ[float64],
) error {
	assm, ok := design.Decl.Assemblies[assembly]
	if !ok {
		return errors.Errorf("couldn't find assembly %s", assembly)
	}
	g, err := assm.Children.PoseDigraph()
	if err != nil {
		return errors.Wrapf(err, "couldn't build scene graph of assembly %s", assembly)
	}
	subtrees := assm.Children.Subtrees()

	nextParents := make([]subtree, 0) // for breadth-first tree traversal
	nextParents = append(nextParents, subtree{id: "", node: root})
	for len(nextParents) > 0 {
		nextParent := nextParents[0]
		nextParents = nextParents[1:]

		// Add leaf nodes first, to get an easier-to-inspect ordering of nodes in the resulting GLTF
		// file:
		intermediateNodes := make([]designs.CompID, 0)
		for _, child := range slices.Sorted(maps.Keys(g[nextParent.id])) {
			if len(subtrees[child].Children) > 0 {
				intermediateNodes = append(intermediateNodes, child)
				continue
			}
			if design.Decl.Components[child].Kind != designs.CompKindPrimitive {
				// This leaf-node component has no models to add to the scene.
				continue
			}

			if _, err = d.addComponent(
				design.FS, child, design.Decl.Components[child], subtrees[child],
				nextParent.node, gridSpacings,
			); err != nil {
				return errors.Wrapf(err, "couldn't add leaf node %s", child)
			}
		}

		// Add non-leaf nodes:
		for _, child := range intermediateNodes {
			var added *gltf.Node
			comp := design.Decl.Components[child]
			if added, err = d.addComponent(
				design.FS, child, comp, subtrees[child], nextParent.node, gridSpacings,
			); err != nil {
				return errors.Wrapf(err, "couldn't add non-leaf node %s", child)
			}
			nextParent := subtree{id: child, node: added}
			nextParents = append(nextParents, nextParent)
		}
	}
	return nil
}

type subtree struct {
	id   designs.CompID
	node *gltf.Node
}

func (d *Document) addComponent(
	designFS ffs.PathedFS, child designs.CompID, comp designs.CompSpec, assmComp designs.AssmCompSpec,
	root *gltf.Node,
	gridSpacings designs.ContinuousXYZ[float64],
) (added *gltf.Node, err error) {
	added = new(gltf.Node)
	added.Name = string(child)
	var nodeIndex int
	d.d.Nodes, nodeIndex = addElem(d.d.Nodes, added)
	root.Children = append(root.Children, nodeIndex)
	if added.Rotation, added.Translation, err = computeNodePose(
		assmComp.Pose, gridSpacings,
	); err != nil {
		return added, errors.Wrapf(err, "couldn't compute pose of component %s", child)
	}

	if comp.Kind != designs.CompKindPrimitive {
		return added, nil
	}

	// Add the node's primitive model to document:
	switch pk := comp.Primitive.Kind; pk {
	default:
		return added, errors.Errorf("unknown model kind for primitive %s: %s", child, pk)
	case "", "static":
		modelHash, nodeIndices, err := d.addComponentPrimitive(designFS, comp.Primitive)
		if err != nil {
			return added, errors.Wrapf(err, "couldn't add primitive component: %+v", comp.Primitive)
		}
		added.Children = append(added.Children, nodeIndices...)
		d.modelInstances[modelHash] = append(d.modelInstances[modelHash], added.Children)
	}
	return added, nil
}

func addElem[T any](arr []T, elem T) (appended []T, idx int) {
	arr = append(arr, elem)
	return arr, len(arr) - 1
}

func computeNodePose(pose designs.AssmCompPoseSpec, gridSpacings designs.ContinuousXYZ[float64]) (
	rot [4]float64, transl [3]float64, err error,
) {
	mat, err := pose.TransfMat(gridSpacings)
	if err != nil {
		return [4]float64{}, [3]float64{}, errors.Wrap(err, "couldn't compute node pose")
	}
	origin := mat.MulVec3(&vec3.Zero)
	const conversion = 0.001 // convert from mm (optikit units) to m (glTF units)
	origin.Scale(conversion)
	return mat.Quaternion(), origin, nil
}

func (d *Document) addComponentPrimitive(fsys ffs.PathedFS, prim designs.CompPrimSpec) (
	modelHash string, rootNodeIndices []int, err error,
) {
	model := prim.StaticModels.GLTF
	if model == "" {
		// Note: we assume that any omission of GLTF models is intentional, because some primitives
		// don't necessarily need to be included in the GLTF render.
		return "", nil, nil
	}
	contents, err := fs.ReadFile(fsys, model)
	if err != nil {
		return "", nil, errors.Wrapf(err, "couldn't read model %s from %s", model, fsys.Path())
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(contents))
	if instances, ok := d.modelInstances[hash]; ok {
		return hash, d.cloneNodeTrees(instances[0]), nil
	}

	m, err := Load(contents)
	if err != nil {
		return hash, nil, errors.Wrapf(err, "couldn't parse glTF/glb model %s", model)
	}
	rootNodes, err := d.addModel(hash, m.d)
	if err != nil {
		return hash, nil, errors.Wrapf(err, "couldn't add model %s", model)
	}
	return hash, rootNodes, nil
}

func (d *Document) cloneNodeTrees(nodeIndices []int) []int {
	clonedIndices := make(map[int]int) // map from node index -> cloned node index
	// NOTE(ethanjli): we add the parent nodes before adding the child nodes, in breadth-first style:
	for _, nodeIndex := range nodeIndices {
		node := *d.d.Nodes[nodeIndex]
		d.d.Nodes, clonedIndices[nodeIndex] = addElem(d.d.Nodes, &node)
	}
	for _, nodeIndex := range nodeIndices {
		d.d.Nodes[clonedIndices[nodeIndex]].Children = d.cloneNodeTrees(d.d.Nodes[nodeIndex].Children)
	}
	return slices.Sorted(maps.Values(clonedIndices))
}

func (d *Document) addModel(hash string, m *gltf.Document) (rootNodeIndices []int, err error) {
	im := newIndexMappings()
	for i, el := range m.Materials {
		d.d.Materials, im.Materials[i] = addElem(d.d.Materials, el)
	}
	if err := d.addModelAccessors(m, im); err != nil {
		return nil, errors.Wrap(err, "couldn't add accessors for model")
	}
	if err := d.addModelMeshes(m.Meshes, im); err != nil {
		return nil, errors.Wrap(err, "couldn't add meshes for model")
	}
	if err := d.addModelNodes(m.Nodes, im); err != nil {
		return nil, errors.Wrap(err, "couldn't add nodes for model")
	}
	d.addModelExtensionsUsed(m.ExtensionsUsed)
	d.indexMappings[hash] = im
	if len(m.Scenes) == 0 {
		// NOTE(ethanjli): We could return maps.Values(im.Nodes), but if they have any internal
		// parent-child relationships then we will invalidate the overall document by making all nodes
		// be the child of a parent node. It's better to just return an error here instead of silently
		// creating broken gltf/glb outputs.
		return nil, errors.Errorf("model %s doesn't specify any scenes with root nodes", hash)
	}
	var sceneIdx int
	if m.Scene != nil {
		sceneIdx = *m.Scene
	}
	scene := m.Scenes[sceneIdx]
	for _, i := range scene.Nodes {
		rootNodeIndices = append(rootNodeIndices, im.Nodes[i])
	}
	return rootNodeIndices, nil
}

func (d *Document) addModelAccessors(m *gltf.Document, im indexMappings) error {
	for i, el := range m.Accessors {
		target := gltf.TargetNone
		if el.BufferView != nil {
			target = m.BufferViews[*el.BufferView].Target
		}
		data, err := modeler.ReadAccessor(m, el, nil)
		if err != nil {
			return errors.Errorf("couldn't read accessor: %+v", el)
		}
		added := modeler.WriteAccessor(d.d, target, data)
		im.Accessors[i] = added
		d.d.Accessors[added].Min = el.Min
		d.d.Accessors[added].Max = el.Max
	}
	return nil
}

func (d *Document) addModelMeshes(m []*gltf.Mesh, im indexMappings) error {
	for i, el := range m {
		for _, pr := range el.Primitives {
			for attr, idx := range pr.Attributes {
				pr.Attributes[attr] = im.Accessors[idx]
			}
			if pr.Indices != nil {
				idx, ok := im.Accessors[*pr.Indices]
				if !ok {
					return errors.Errorf("couldn't re-map accessor index for mesh primitive: %+v", pr)
				}
				pr.Indices = &idx
			}
			if pr.Material != nil {
				idx, ok := im.Materials[*pr.Material]
				if !ok {
					return errors.Errorf("couldn't re-map material index for mesh primitive: %+v", pr)
				}
				pr.Material = &idx
			}
		}
		d.d.Meshes, im.Meshes[i] = addElem(d.d.Meshes, el)
	}
	return nil
}

func (d *Document) addModelNodes(m []*gltf.Node, im indexMappings) error {
	// NOTE(ethanjli): we add the parent nodes before adding the child nodes, in breadth-first style:
	addedNodes := make(map[int]*gltf.Node, 0)
	for i, el := range m {
		if el.Mesh != nil {
			idx, ok := im.Meshes[*el.Mesh]
			if !ok {
				return errors.Errorf("couldn't re-map mesh index for node: %+v", el)
			}
			el.Mesh = &idx
		}

		addedNodes[i] = el
		d.d.Nodes, im.Nodes[i] = addElem(d.d.Nodes, el)
	}

	for _, node := range addedNodes {
		for i, j := range node.Children {
			idx, ok := im.Nodes[j]
			if !ok {
				return errors.Errorf("couldn't re-map child node index %d for node: %+v", j, *node)
			}
			node.Children[i] = idx
		}
	}
	return nil
}

func (d *Document) addModelExtensionsUsed(me []string) {
	eu := make(structures.Set[string])
	for _, e := range d.d.ExtensionsUsed {
		eu.Add(e)
	}
	for _, e := range me {
		eu.Add(e)
	}
	d.d.ExtensionsUsed = slices.Sorted(eu.All())
}

func (d *Document) Encode(asText bool) ([]byte, error) {
	var buf bytes.Buffer
	enc := gltf.NewEncoder(&buf)
	enc.AsBinary = !asText
	if err := enc.Encode(d.d); err != nil {
		return nil, err
	}
	if asText {
		result := buf.Bytes()
		buf.Reset()
		if err := json.Indent(&buf, result, "", "  "); err != nil {
			return result, err
		}
	}
	return buf.Bytes(), nil
}

// indexMappings

func newIndexMappings() indexMappings {
	return indexMappings{
		Accessors: make(map[int]int),
		Materials: make(map[int]int),
		Meshes:    make(map[int]int),
		Nodes:     make(map[int]int),
	}
}
