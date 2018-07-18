package installer

import (
	"io"
	"os"

	"go.evanpurkhiser.com/dots/config"
	"go.evanpurkhiser.com/dots/resolver"
)

func OpenDotfile(dotfile *resolver.Dotfile, config config.SourceConfig) (io.ReadCloser, error) {
	files := make([]*os.File, len(dotfile.Sources))

	for i, source := range dotfile.Sources {
		file, err := os.Open(config.SourcePath + separator + source.Path)
		if err != nil {
			return nil, err
		}

		files[i] = file
	}

	compiler := &dotfileCompiler{
		dotfile: dotfile,
		config:  config,
		files:   files,
	}

	return compiler, nil
}

type dotfileCompiler struct {
	dotfile *resolver.Dotfile
	config  config.SourceConfig
	files   []*os.File
}

// TODO: If we want we can make this thing do caching of compiled dotfiels if
// we expect to install them later
func (c *dotfileCompiler) Read(p []byte) (n int, err error) {
	readers := []io.Reader{}

	for _, file := range c.files {
		readers = append(readers, file)
	}

	// TODO: Implement filtered reading
	return io.MultiReader(readers...).Read(p)
}

func (c *dotfileCompiler) Close() error {
	var err error

	for _, file := range c.files {
		closeErr := file.Close()
		if closeErr != nil {
			err = closeErr
		}
	}

	return err
}
