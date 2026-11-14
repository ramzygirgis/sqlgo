package parsing

import (
	"fmt"

	"github.com/ramzygirgis/sqlgo/internal/lexing"
)

type Cardinality int

const (
	IllegalCardinality Cardinality = iota
	One
	Many
	Exec
)

func (c Cardinality) String() string {
	strings := [...]string{
		"IllegalCardinality", "One", "Many", "Exec",
	}
	if c < 0 || int(c) >= len(strings) {
		return fmt.Sprintf("Cardinality(%d)", c)
	}
	return strings[c]
}

type Annotation struct {
	Name        string
	Cardinality Cardinality
}

type AnnotatedQuery struct {
	Annotation Annotation
	Statement  *Statement
}

func (p *parser) expectAnnotation() (Annotation, error) {
	tok := p.read()
	switch tok.Kind {
	case lexing.OneAnnotation:
		p.next()
		return Annotation{
			Name:        tok.Body,
			Cardinality: One,
		}, nil
	case lexing.ManyAnnotation:
		p.next()
		return Annotation{
			Name:        tok.Body,
			Cardinality: Many,
		}, nil
	case lexing.ExecAnnotation:
		p.next()
		return Annotation{
			Name:        tok.Body,
			Cardinality: Exec,
		}, nil
	default:
		return Annotation{}, fmt.Errorf("expected tokenKind to be for an annotation, got %s", tok.Kind)
	}
}

func (p *parser) ParseQuery() (*AnnotatedQuery, error) {
	annotation, err := p.expectAnnotation()
	if err != nil {
		return nil, err
	}
	statement, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	return &AnnotatedQuery{
		Annotation: annotation,
		Statement:  &statement,
	}, nil
}

func ParseQueries(query string) ([]*AnnotatedQuery, error) {
	toks, err := lexing.Lex(query)
	if err != nil {
		return nil, err
	}
	if len(toks) == 0 {
		return nil, fmt.Errorf("empty token stream")
	}
	p := parser{
		toks: toks,
		pos:  0,
	}
	var seenFunctionNames map[string]struct{}
	var annotatedQueries []*AnnotatedQuery
	for p.read().Kind != lexing.EOF && p.read().Kind != lexing.OutOfBounds {
		query, err := p.ParseQuery()
		if err != nil {
			return nil, err
		}
		if _, ok := seenFunctionNames[query.Annotation.Name]; ok {
			return nil, fmt.Errorf("query %q appears more than once", query.Annotation.Name)
		}
		annotatedQueries = append(annotatedQueries, query)
		seenFunctionNames[query.Annotation.Name] = struct{}{}
	}

	return annotatedQueries, nil
}
