package ssml

import (
	"regexp"
	"strings"
)

type SSMLAttribute struct {
	name  string
	value string
}

type SSMLNode struct {
	attributes []SSMLAttribute
	value      string
	children   []*SSMLNode
}

func newNode(value string, attributes []SSMLAttribute) *SSMLNode {
	return &SSMLNode{attributes: attributes, children: []*SSMLNode{}, value: value}
}

type Parser interface {
	ParseTree(input string) *SSMLNode
	tokenize(input string) []Token
}

type Token struct {
	value string
}

type Tokenizer interface {
	tokenize(input string) []Token
	tokenizeAttributes(input string) []Token
}

type WordTokenizer struct{}

type SpecialStr string
type SpecialChar rune

const (
	SpecialCharNewLine        SpecialChar = '\n'
	SpecialCharTab            SpecialChar = '\t'
	SpecialCharCarriageReturn SpecialChar = '\r'
	SpecialCharLessThan       SpecialChar = '<'
	SpecialCharGreaterThan    SpecialChar = '>'
	SpecialCharAmpersand      SpecialChar = '&'
)

const (
	SpecialStrLessThan       SpecialStr = "&lt;"
	SpecialStrGreaterThan    SpecialStr = "&gt;"
	SpecialStrAmpersand      SpecialStr = "&amp;"
	SpecialStrClosingTag     SpecialStr = "</"
	SpecialStrSelfClosingTag SpecialStr = "/>"
)

func (t *Token) isTag() bool {
	return t.isClosingTag() || t.isSelfClosingTag() || t.isOpeningTag()
}

func (t *Token) isSelfClosingTag() bool {
	return strings.HasPrefix(t.value, string(SpecialCharLessThan)) && strings.HasSuffix(t.value, string(SpecialStrSelfClosingTag))
}

func (t *Token) isClosingTag() bool {
	return strings.HasPrefix(t.value, string(SpecialStrClosingTag)) && strings.HasSuffix(t.value, string(SpecialCharGreaterThan))
}

func (t *Token) isOpeningTag() bool {
	return !t.isClosingTag() && !t.isSelfClosingTag() && strings.HasPrefix(t.value, string(SpecialCharLessThan)) && strings.HasSuffix(t.value, string(SpecialCharGreaterThan))
}

func (t *Token) extractValue() string {
	if !t.isTag() {
		return t.value
	}
	s := strings.ReplaceAll(t.value, string(SpecialStrClosingTag), "")
	s = strings.ReplaceAll(s, string(SpecialStrSelfClosingTag), "")
	s = strings.ReplaceAll(s, string(SpecialCharLessThan), "")
	s = strings.ReplaceAll(s, string(SpecialCharGreaterThan), "")
	s = strings.TrimSpace(s)
	s = strings.Split(s, " ")[0]
	return s
}

func (wordTokenizer *WordTokenizer) tokenize(input string) []Token {
	regex := regexp.MustCompile(`<[a-zA-Z0-9_/=\-" .:]*>|( *[^<>\s][^<>]*)`)

	s := strings.ReplaceAll(input, string(SpecialCharTab), "")
	s = strings.ReplaceAll(s, string(SpecialCharNewLine), "")
	s = strings.ReplaceAll(s, string(SpecialCharCarriageReturn), "")
	s = strings.ReplaceAll(s, string(SpecialStrLessThan), string(SpecialCharLessThan))
	s = strings.ReplaceAll(s, string(SpecialStrGreaterThan), string(SpecialCharGreaterThan))
	s = strings.ReplaceAll(s, string(SpecialStrAmpersand), string(SpecialCharAmpersand))
	s = strings.TrimSpace(s)

	matches := regex.FindAllString(s, -1)

	tokens := []Token{}

	for _, value := range matches {
		tokens = append(tokens, Token{value: value})
	}

	return tokens
}

func (wordTokenizer *WordTokenizer) tokenizeAttributes(input string) []Token {
	regex := regexp.MustCompile(`[:a-zA-Z0-9]+ *= *"[a-zA-Z0-9 .-]*"`)

	s := strings.ReplaceAll(input, string(SpecialCharTab), "")
	s = strings.ReplaceAll(s, string(SpecialCharNewLine), "")
	s = strings.ReplaceAll(s, string(SpecialCharCarriageReturn), "")
	s = strings.TrimSpace(s)

	matches := regex.FindAllString(s, -1)

	tokens := []Token{}

	for _, value := range matches {
		tokens = append(tokens, Token{value: value})
	}

	return tokens
}

func NewWordTokenizer() Tokenizer {
	return &WordTokenizer{}
}

type SSMLParser struct {
	Tokenizer
}

func NewSSMLParser(tokenizer Tokenizer) Parser {
	return &SSMLParser{Tokenizer: tokenizer}
}

func (parser *SSMLParser) ParseTree(input string) *SSMLNode {
	tokens := parser.tokenize(input)
	var maxParsed *int = new(int)
	*maxParsed = 0
	root := parser.addNode(nil, 0, tokens, maxParsed)
	//root := parser.constructSSMLTree(tokens)
	maxParsed = nil
	return root
}

func (parser *SSMLParser) ParseAttributes(input string) []SSMLAttribute {
	tokens := parser.tokenizeAttributes(input)
	attributes := make([]SSMLAttribute, 0)
	for _, token := range tokens {
		split := strings.Split(token.value, "=")
		name, value := split[0], split[1]
		attributes = append(attributes, SSMLAttribute{name: strings.TrimSpace(name), value: strings.ReplaceAll(value, `"`, "")})
	}
	return attributes
}

func (parser *SSMLParser) getNodeFromToken(token Token) *SSMLNode {
	return newNode(token.extractValue(), parser.ParseAttributes(token.value))
}

func (parser *SSMLParser) addNode(parent *SSMLNode, index int, tokens []Token, maxParsed *int) *SSMLNode {
	if index >= len(tokens) {
		return parent
	}
	*maxParsed = max(*maxParsed, index)
	currentToken := tokens[index]

	if parent == nil {
		newParent := parser.getNodeFromToken(currentToken)
		return parser.addNode(newParent, index+1, tokens, maxParsed)
	}

	if currentToken.isClosingTag() {
		return parent
	}

	// token has been parsed
	if index < *maxParsed {
		return parser.addNode(parent, *maxParsed+1, tokens, maxParsed)
	}

	var nodeToAdd *SSMLNode
	if currentToken.isOpeningTag() {
		newParent := parser.getNodeFromToken(currentToken)
		nodeToAdd = parser.addNode(newParent, index+1, tokens, maxParsed)
	} else {
		nodeToAdd = parser.getNodeFromToken(currentToken)
	}

	parent.children = append(parent.children, nodeToAdd)

	return parser.addNode(parent, index+1, tokens, maxParsed)
}

/* loop form */
func (parser *SSMLParser) constructSSMLTree(tokens []Token) *SSMLNode {
	var root *SSMLNode
	var current *SSMLNode
	stack := []*SSMLNode{}

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]

		if token.isClosingTag() {
			if len(stack) > 0 {
				current = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
			}
			continue
		}

		node := parser.getNodeFromToken(token)

		if root == nil {
			root = node
		}

		if current != nil {
			current.children = append(current.children, node)
		}

		if token.isOpeningTag() {
			stack = append(stack, current)
			current = node
		}
	}

	return root
}
