package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
	LocalScope  SymbolScope = "LOCAL"
)

// Symbol holds all the necessary information about a symbol we encounter.
type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

// SymbolTable associates the identifiers we come across with Symbols in a
// map (store) and keeps track of the number of definitions it has.
type SymbolTable struct {
	Outer    *SymbolTable
	store    map[string]Symbol
	defCount int
}

// NewSymbolTable returns a pointer to a new instance of SymbolTable.
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store: make(map[string]Symbol),
	}
}

// NewEnclosedSymbolTable returns a pointer to a new instance of
// SymbolTable linked to the outer symbol-table passed, i.e., enclosed
// within the outer.
func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

// Define creates a new Symbol with the given name, assigns it the next available
// index, and stores it in the symbol table. Returns the newly created Symbol.
func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Index: s.defCount}
	if s.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}
	s.store[name] = symbol
	s.defCount++
	return symbol
}

// Resolve looks up a symbol by name in the symbol table. Returns the Symbol
// and true if found, or an empty Symbol and false if not found.
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	symbol, ok := s.store[name]
	if !ok && s.Outer != nil {
		return s.Outer.Resolve(name)
	}
	return symbol, ok
}
