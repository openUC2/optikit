package optikit

import (
	"context"

	"github.com/openUC2/optikit/internal/clients/build123d"
)

// Model

func ConvertModel(
	ctx context.Context, inputFormat, inputPath, outputFormat, outputPath string,
) error {
	cqc, err := build123d.New()
	if err != nil {
		return err
	}
	defer func() {
		err = cqc.Close()
	}()

	if err = cqc.Convert(inputFormat, inputPath, outputFormat, outputPath); err != nil {
		return err
	}
	return nil
}
