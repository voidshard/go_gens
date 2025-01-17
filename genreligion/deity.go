package genreligion

import "github.com/Flokey82/go_gens/genlanguage"

var DeityMeaningApproaches []string

func init() {
	DeityMeaningApproaches = weightedToArray(GenMeaningApproaches)
}

// Deity represents a deity name.
type Deity struct {
	Name     string
	Meaning  string
	Approach string
}

// FullName returns the full name of the deity (including the meaning, if any).
func (d *Deity) FullName() string {
	if d == nil {
		return ""
	}
	if d.Meaning == "" {
		return d.Name
	}
	return d.Name + ", The " + d.Meaning
}

// GetDeity returns a deity name for the given culture.
// This code is based on:
// https://github.com/Azgaar/Fantasy-Map-Generator/blob/master/modules/religions-generator.js
func (g *Generator) GetDeity(lang *genlanguage.Language, approach string) *Deity {
	if lang == nil {
		return nil
	}
	if approach == "" {
		approach = g.RandDeityGenMethod()
	}
	return &Deity{
		Name:     lang.MakeName(),
		Meaning:  g.GenerateDeityMeaning(approach),
		Approach: approach,
	}
}
