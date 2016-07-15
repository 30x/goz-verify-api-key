package pipeline

import (
	"io"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"fmt"
	"errors"
)

// Definition of a pipe
type Definition interface {
	CreatePipe(reqID string) Pipe
}

// DefinePipeFromReader returns a Pipe Definition as defined in the yaml Reader
func DefinePipe(r io.Reader) (Definition, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	type FittingDef map[string]interface{}
	type PipeDef map[string][]FittingDef

	var pipeConfig PipeDef
	err = yaml.Unmarshal(bytes, &pipeConfig)
	if err != nil {
		return nil, err
	}

	if pipeConfig["request"] == nil && pipeConfig["response"] == nil {
		return nil, errors.New("Illegal Definition: Must have 'request' and/or 'response' keys")
	}

	// todo: validation of structures

	var reqFittings []FittingWithID
	reqDefs := pipeConfig["request"]
	if reqDefs != nil {
		for _, fittingDef := range reqDefs {
			if len(fittingDef) > 1 {
				return nil, fmt.Errorf("bad structure")
			}
			for id, config := range fittingDef{
				fitting, err := NewFitting(id, config)
				if err != nil {
					return nil, err
				}
				handler := fitting.RequestHandlerFunc()
				if handler != nil {
					reqFittings = append(reqFittings, fitting)
				}
			}
		}
	}

	var resFittings []FittingWithID
	resDefs := pipeConfig["response"]
	if resDefs != nil {
		for _, fittingDef := range resDefs {
			if len(fittingDef) > 1 {
				return nil, fmt.Errorf("bad structure")
			}
			for id, config := range fittingDef{
				fitting, err := NewFitting(id, config)
				if err != nil {
					return nil, err
				}
				handler := fitting.ResponseHandlerFunc()
				if handler != nil {
					resFittings = append(resFittings, fitting)
				}
			}
		}
	}

	return NewDefinition(reqFittings, resFittings), nil
}

// DefinePipe returns a Pipe Definition defined by the passed handlers
func NewDefinition(reqFittings []FittingWithID, resFittings []FittingWithID) Definition {
	return &definition{reqFittings, resFittings}
}

type definition struct {
	reqFittings []FittingWithID
	resFittings []FittingWithID
}

// if reqId is nil, will create and use an internal id
func (d *definition) CreatePipe(reqID string) Pipe {
	return newPipe(reqID, d.reqFittings, d.resFittings)
}
