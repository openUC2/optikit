package build123d

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kluctl/go-embed-python/embed_util"
	"github.com/kluctl/go-embed-python/python"

	"github.com/openUC2/optikit/internal/clients/build123d/data"
)

type Client struct {
	epy  *python.EmbeddedPython
	epip *embed_util.EmbeddedFiles
	esrc *embed_util.EmbeddedFiles
}

func New() (c *Client, err error) {
	c = &Client{}
	if c.epy, err = python.NewEmbeddedPython("cadquery"); err != nil {
		return nil, err
	}

	if c.epip, err = embed_util.NewEmbeddedFiles(data.Data, "cadquery"); err != nil {
		return nil, err
	}
	c.epy.AddPythonPath(c.epip.GetExtractedPath())

	if c.esrc, err = embed_util.NewEmbeddedFiles(Source, "cad"); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) Close() error {
	return errors.Join(c.epy.Cleanup(), c.epip.Cleanup())
}

func (c *Client) Run() ([]byte, error) {
	cmd, err := c.epy.PythonCmd(filepath.Join(c.esrc.GetExtractedPath(), "main.py"))
	if err != nil {
		return nil, err
	}

	// TODO: write the script to a temporary file instead!
	in := bytes.Buffer{}
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
