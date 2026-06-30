package build123d

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kluctl/go-embed-python/embed_util"
	"github.com/kluctl/go-embed-python/pip"
	"github.com/kluctl/go-embed-python/python"

	"github.com/openUC2/optikit/internal/clients/build123d/data"
)

type Client struct {
	pyEmbed   *python.EmbeddedPython
	pipEmbeds []*embed_util.EmbeddedFiles
	srcEmbed  *embed_util.EmbeddedFiles
}

func New() (c *Client, err error) {
	c = &Client{}
	if c.pyEmbed, err = python.NewEmbeddedPython("cadquery"); err != nil {
		return nil, err
	}

	epip, err := embed_util.NewEmbeddedFiles(data.Data, "cadquery")
	if err != nil {
		return nil, err
	}
	c.addEmbeddedPipLib(epip)
	c.pyEmbed.AddPythonPath(epip.GetExtractedPath())
	c.pipEmbeds = append(c.pipEmbeds, epip)

	if c.srcEmbed, err = embed_util.NewEmbeddedFiles(Source, "cad"); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) addEmbeddedPipLib(pipEmbed *embed_util.EmbeddedFiles) {
	c.pyEmbed.AddPythonPath(pipEmbed.GetExtractedPath())
	c.pipEmbeds = append(c.pipEmbeds, pipEmbed)
}

func (c *Client) Close() error {
	errs := []error{c.pyEmbed.Cleanup()}
	for _, epip := range c.pipEmbeds {
		if err := epip.Cleanup(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// PipFreeze returns the result of running `pip freeze` with the requirements.txt file at the
// specified path.
// Note: this has a side-effect of adding an embedded pip library to the embedded Python instance's
// Python path.
func (c *Client) PipFreeze(requirementsFile string) (result []byte, err error) {
	pipLib, err := pip.NewPipLib("pip")
	if err != nil {
		return nil, err
	}
	c.addEmbeddedPipLib(pipLib)

	cmd, err := c.pyEmbed.PythonCmd("-m", "pip", "freeze", "-r", requirementsFile)
	if err != nil {
		return nil, err
	}

	out := bytes.Buffer{}
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		fmt.Println(out.String())
		return nil, err
	}
	return out.Bytes(), err
}

func (c *Client) Assemble(stdin []byte) ([]byte, error) {
	cmd, err := c.pyEmbed.PythonCmd(filepath.Join(c.srcEmbed.GetExtractedPath(), "assemble.py"))
	if err != nil {
		return nil, err
	}

	in := bytes.Buffer{}
	in.Write(stdin)
	cmd.Stdin = &in

	out := bytes.Buffer{}
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err = cmd.Run(); err != nil {
		fmt.Println(out.String())
		return nil, err
	}
	return out.Bytes(), err
}

func (c *Client) Convert(inputFormat, inputPath, outputFormat, outputPath string) error {
	cmd, err := c.pyEmbed.PythonCmd(
		filepath.Join(c.srcEmbed.GetExtractedPath(), "convert.py"),
		inputFormat, inputPath, outputFormat, outputPath,
	)
	if err != nil {
		return err
	}

	if err = cmd.Run(); err != nil {
		return err
	}
	return nil
}
