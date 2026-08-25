package gltf

import (
	"bytes"
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
		d: gltf.NewDocument(),
		root: &gltf.Node{
			Name: "design",
		},
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
	design *designs.FSDesign, asText bool, gridSpacings designs.ContinuousXYZ[float64],
) ([]byte, error) {
	if err := d.addComponents(design, gridSpacings, d.root); err != nil {
		return nil, err
	}

	if asText {
		return d.Encode(true)
	}

	return d.Encode(false)
}

func (d *Document) addComponents(
	design *designs.FSDesign, gridSpacings designs.ContinuousXYZ[float64], root *gltf.Node,
) error {
	flattened := design.Decl.Components.TranslFlattened()
	compIDs := slices.Sorted(maps.Keys(flattened))
	subdesignCompIDs := make([]designs.CompID, 0, len(compIDs))
	for _, id := range compIDs {
		comp := flattened[id]
		if comp.Pose.Rotation.Type == "" {
			continue
		}

		switch t := comp.Type; t {
		default:
			return errors.Errorf("unknown component type for component %s: %s", id, t)
		case designs.CompTypeLocation:
			return nil
		case designs.CompTypePrimitive:
			if err := d.addPrimitiveComponent(design.FS, id, comp, gridSpacings, root); err != nil {
				return errors.Wrapf(err, "couldn't add primitive component %s to gltf model", id)
			}
		case designs.CompTypeDesign:
			// We recurse into subdesigns only after adding all other nodes, in order to construct the
			// glTF scene graph in breadth-first order instead of depth-first order:
			subdesignCompIDs = append(subdesignCompIDs, id)
		}
	}
	for _, id := range subdesignCompIDs {
		comp := flattened[id]
		subdesign, err := design.LoadCompFSDesign(id)
		if err != nil {
			return errors.Wrapf(
				err, "couldn't load subdesign %s for component %s", comp.Design, id,
			)
		}
		if err := d.addSubdesignComponent(id, comp, subdesign, gridSpacings, root); err != nil {
			return errors.Wrapf(err, "couldn't add subdesign component %s to gltf model", id)
		}
	}
	return nil
}

func (d *Document) addPrimitiveComponent(
	fsys ffs.PathedFS, id designs.CompID, comp designs.CompSpec,
	gridSpacings designs.ContinuousXYZ[float64], parent *gltf.Node,
) error {
	n := new(gltf.Node)
	n.Name = string(id)
	var nodeIndex int
	d.d.Nodes, nodeIndex = addElem(d.d.Nodes, n)
	parent.Children = append(parent.Children, nodeIndex)

	var err error
	if n.Rotation, n.Translation, err = computeNodePose(comp.Pose, gridSpacings); err != nil {
		return errors.Wrapf(err, "couldn't compute pose of component %s", id)
	}

	// Add the node's primitive model to document:
	switch pt := comp.Primitive.Type; pt {
	default:
		return errors.Errorf("unknown model type for primitive %s: %s", id, pt)
	case "", "static":
		modelHash, nodeIndices, err := d.addComponentPrimitive(fsys, comp.Primitive)
		if err != nil {
			return errors.Wrapf(err, "couldn't add primitive component: %+v", comp.Primitive)
		}
		n.Children = append(n.Children, nodeIndices...)
		d.modelInstances[modelHash] = append(d.modelInstances[modelHash], n.Children)
	}
	return nil
}

func addElem[T any](arr []T, elem T) (appended []T, idx int) {
	arr = append(arr, elem)
	return arr, len(arr) - 1
}

func computeNodePose(pose designs.CompPoseSpec, gridSpacings designs.ContinuousXYZ[float64]) (
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
		return "", nil, errors.Errorf("primitive component has no glTF/glb model file: %+v", prim)
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

func (d *Document) addSubdesignComponent(
	id designs.CompID, comp designs.CompSpec, subdesign *designs.FSDesign,
	gridSpacings designs.ContinuousXYZ[float64], parent *gltf.Node,
) error {
	n := new(gltf.Node)
	n.Name = string(id)
	var nodeIndex int
	d.d.Nodes, nodeIndex = addElem(d.d.Nodes, n)
	parent.Children = append(parent.Children, nodeIndex)

	var err error
	if n.Rotation, n.Translation, err = computeNodePose(comp.Pose, gridSpacings); err != nil {
		return errors.Wrapf(err, "couldn't compute pose of component %s", id)
	}

	if err = d.addComponents(subdesign, gridSpacings, n); err != nil {
		return errors.Wrapf(
			err, "couldn't add subcomponents of component %s with design %s", id, comp.Design,
		)
	}
	return nil
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
