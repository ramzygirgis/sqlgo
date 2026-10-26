package parsing

import (
	"github.com/ramzygirgis/sqlgo/internal/lexing"
	"strings"
)

type CreateTableStatement struct {
	TableName   string
	Columns     []Column
	IfNotExists bool
}

func (c *CreateTableStatement) statementMarker() {}

func (c *CreateTableStatement) String() string {
	var sb strings.Builder

	sb.WriteString("CREATE TABLE ")
	if c.IfNotExists {
		sb.WriteString("IF NOT EXISTS ")
	}
	sb.WriteString(c.TableName)
	sb.WriteString(" (\n")

	colStrs := make([]string, len(c.Columns))
	for i, col := range c.Columns {
		colStrs[i] = "   " + col.String()
	}
	sb.WriteString(strings.Join(colStrs, ",\n"))

	sb.WriteString("\n);")
	return sb.String()
}

type Column struct {
	Name       string
	SqlType    string
	HasDefault bool
	DefaultVal string
	Nullable   bool
	PrimaryKey bool
}

func (col Column) String() string {
	parts := []string{col.Name, col.SqlType}
	if !col.Nullable {
		parts = append(parts, "NOT NULL")
	}
	if col.PrimaryKey {
		parts = append(parts, "PRIMARY KEY")
	}
	if col.HasDefault {
		parts = append(parts, "DEFAULT "+col.DefaultVal)
	}
	return strings.Join(parts, " ")
}

func (p *parser) parseCreateTable() (*CreateTableStatement, error) {
	ifNotExists := false
	if _, err := p.expect(lexing.Keyword, "CREATE"); err != nil {
		return nil, err
	}
	if _, err := p.expect(lexing.Keyword, "TABLE"); err != nil {
		return nil, err
	}
	if (p.read() == lexing.Token{
		Kind: lexing.Keyword,
		Body: "IF",
	}) {
		p.next()
		if err := p.expectKeyword("NOT"); err != nil {
			return nil, err
		}
		if err := p.expectKeyword("EXISTS"); err != nil {
			return nil, err
		}
		ifNotExists = true
	}
	tableName, err := p.expectKind(lexing.Identifier)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexing.LParen, "("); err != nil {
		return nil, err
	}
	var cols []Column
	col, err := p.parseColumnDef()
	if err != nil {
		return nil, err
	}
	cols = append(cols, col)
	for p.read().Kind == lexing.Comma {
		p.next()
		col, err := p.parseColumnDef()
		if err != nil {
			return nil, err
		}
		cols = append(cols, col)
	}
	if _, err = p.expect(lexing.RParen, ")"); err != nil {
		return nil, err
	}
	if _, err := p.expect(lexing.Semicolon, ";"); err != nil {
		return nil, err
	}
	return &CreateTableStatement{
		TableName:   tableName,
		Columns:     cols,
		IfNotExists: ifNotExists,
	}, err
}

func (p *parser) parseColumnDef() (Column, error) {
	name, err := p.expectKind(lexing.Identifier)
	nullable := true
	primaryKey := false
	hasDefault := false
	defaultVal := ""
	if err != nil {
		return Column{}, err
	}
	sqlType, err := p.expectKind(lexing.Type)
	if err != nil {
		return Column{}, err
	}
constraints:
	for p.read().Kind != lexing.EOF { // find cleanest way to do this
		switch p.read() {
		case lexing.Token{
			Kind: lexing.Keyword,
			Body: "NOT",
		}:
			p.next()
			err := p.expectKeyword("NULL")
			if err != nil {
				return Column{}, err
			}
			nullable = false
		case lexing.Token{
			Kind: lexing.Keyword,
			Body: "PRIMARY",
		}:
			p.next()
			err := p.expectKeyword("KEY")
			if err != nil {
				return Column{}, err
			}
			primaryKey = true
		case lexing.Token{
			Kind: lexing.Keyword,
			Body: "DEFAULT",
		}:
			p.next()
			body, err := p.expectKind(lexing.Identifier)
			if err != nil {
				return Column{}, err
			}
			hasDefault = true
			defaultVal = body
		default:
			break constraints
		}
	}
	return Column{
		Name:       name,
		SqlType:    sqlType,
		HasDefault: hasDefault,
		DefaultVal: defaultVal,
		Nullable:   nullable,
		PrimaryKey: primaryKey,
	}, nil
}
