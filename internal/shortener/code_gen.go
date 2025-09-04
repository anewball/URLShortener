package shortener

import gonanoid "github.com/matoous/go-nanoid/v2"

type NanoID interface {
	Generate(n int) (string, error)
}

type nanoIDImpl struct {
	alphabet string
}

func NewNanoID(alphabet string) NanoID {
	return &nanoIDImpl{alphabet: alphabet}
}

func (c *nanoIDImpl) Generate(n int) (string, error) {
	return gonanoid.Generate(c.alphabet, n)
}
