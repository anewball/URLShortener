package shortener

import gonanoid "github.com/matoous/go-nanoid/v2"

type NanoID interface {
	Generate(n int) (string, error)
}

type nanoID struct {
	alphabet string
}

var _ NanoID = (*nanoID)(nil)

func NewNanoID(alphabet string) NanoID {
	return &nanoID{alphabet: alphabet}
}

func (c *nanoID) Generate(n int) (string, error) {
	return gonanoid.Generate(c.alphabet, n)
}
