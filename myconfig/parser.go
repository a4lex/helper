package myconfig

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

const (
	tkPostKey      = ":"
	tkPreListValue = "- "
	tkComment      = "#"
	tkStringDelim1 = "\""
	tkStringDelim2 = "'"
)

// configParser is the struct for parsing YAML files
type configParser struct {
	reader     *bufio.Reader
	readCursor uint   // Current read cursor position (on current line)
	readBytes  []byte // Current line read bytes

	rootNode      Node
	currentNode   *Node
	indentSetting uint // Global indentation setting

	prevIndent uint // Previous indentation
	line       uint // Current line
}

// NewParser returns a new configParser to be used for YAML parsing.
// file YAML file to parse
func NewParser(reader *bufio.Reader) *configParser {

	yp := new(configParser)
	yp.reader = reader

	yp.rootNode = createRootNode()
	yp.currentNode = &yp.rootNode

	return yp
}

// Parse returns a YAML (root node + children nodes) from the input file.
func (yp *configParser) Parse() (*Node, error) {
	var err error

	yp.line = 1

	for {
		err = yp.readLineBytes()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		if len(yp.readBytes) > 0 {
			err = yp.parseLine()
			if err != nil {
				return nil, err
			}
		}

		yp.line++
	}

	return &yp.rootNode, nil
}

// readLineBytes sets the read bytes until the new line character
// into the Parser struct.
func (yp *configParser) readLineBytes() error {
	var c uint

	// Clear read
	yp.readCursor = 0
	yp.readBytes = yp.readBytes[:0]

	for {
		b, err := yp.reader.ReadByte()

		if err != nil {
			return err
		}

		if b == '\n' {
			c = uint(len(yp.readBytes))

			// Handle Windows' Line Endings
			if c > 1 && yp.readBytes[c-1] == '\r' {
				yp.readBytes = yp.readBytes[0 : c-1]
			}
			break
		}

		yp.readBytes = append(yp.readBytes, b)
	}

	return nil
}

// pick reads one character from the read bytes.
func (yp *configParser) pick() string {
	return string(yp.readBytes[yp.readCursor])
}

// prev moves the cursor one step forward.
func (yp *configParser) next() bool {
	if int(yp.readCursor) == len(yp.readBytes)-1 {
		return false
	}

	yp.readCursor++

	return true
}

// consumeSpaces returns the X read spaces from the read bytes
// and stops at the first non-space character.
func (yp *configParser) consumeSpaces() uint {
	var c uint = 0

	for {
		char := yp.pick()
		if char != " " {
			break
		}

		c++

		if !yp.next() {
			break
		}
	}

	return c
}

// determineIndent returns the indent level for the given raw indent.
func (yp *configParser) determineIndent(rawIndent uint) (uint, error) {
	if rawIndent == 0 {
		return 0, nil
	}

	if yp.indentSetting == 0 && rawIndent > 0 {
		// Initialize indent setting
		yp.indentSetting = rawIndent
	}

	if rawIndent%yp.indentSetting > 0 {
		err := yp.err("Syntax Error! Invalid indent")
		return 0, err
	}

	return rawIndent / yp.indentSetting, nil
}

// parseLine constructs the YAML tree by parsing the read bytes.
func (yp *configParser) parseLine() error {
	var rawIndent = yp.consumeSpaces()

	if yp.pick() == tkComment {
		// Nothing but comment, skip
		return nil
	}

	indent, err := yp.determineIndent(rawIndent)

	if err != nil {
		return err
	}

	// Check for value type many
	if indent >= yp.prevIndent && yp.isListType() {
		yp.getListValue()
		return nil
	}

	// From now on it can only be a new node

	indentShift := int(indent) - int(yp.prevIndent)

	if indentShift > 1 {
		return yp.err("Syntax Error! Invalid indent (no parent)")
	}

	parentNodePtr := yp.currentNode

	if indentShift <= 0 && yp.currentNode.parent != nil {
		// Rewind to correct node
		for i := 0; i < (-indentShift)+1; i++ {
			parentNodePtr = parentNodePtr.parent
		}
	}

	yp.currentNode = newChildNode(parentNodePtr)

	yp.processKey()

	if yp.move(len(tkPostKey)) {
		yp.consumeSpaces()
		yp.processValue()
	}

	yp.prevIndent = indent

	return nil
}

func (yp *configParser) processKey() error {
	k, err := yp.parseValue(tkPostKey)

	if err != nil {
		return err
	}

	if k == "" {
		err := yp.err("Syntax Error! Key can't be null")
		return err
	}

	yp.currentNode.name = k

	return nil
}

func (yp *configParser) processValue() error {
	v, err := yp.parseValue("")

	if err != nil {
		return err
	}

	if v == "" {
		yp.currentNode.ntype = nodeTypeChildren
	} else {
		yp.currentNode.ntype = nodeTypeScalar
		yp.appendValue(v)
	}

	return nil
}

// isListType tells if the read bytes is type-of list.
func (yp *configParser) isListType() bool {
	return yp.equals(tkPreListValue)
}

// getListValue parses the read bytes to get the current list value.
func (yp *configParser) getListValue() error {
	// Skip prefix token
	yp.move(len(tkPreListValue))

	v, err := yp.parseValue("")

	if err != nil {
		return err
	}

	yp.currentNode.ntype = nodeTypeList
	yp.appendValue(v)

	return nil
}

// appendValue appends the given value to the current node.
func (yp *configParser) appendValue(value string) {
	yp.currentNode.values = append(yp.currentNode.values, value)
}

// equals tells if the given string matches the beginning of the read bytes.
func (yp *configParser) equals(str string) bool {
	return yp.read(len(str)) == str
}

// parseValue returns the parsed value from the read bytes.
func (yp *configParser) parseValue(stopChar string) (string, error) {
	var value string
	inString := false
	var strDelim string

	for {
		char := yp.pick()

		if stopChar != "" && char == stopChar {
			break
		}

		if !inString && char == tkComment {
			// Comments are only valid if preceded by a space
			if yp.read(-1) == " " {
				// Move at the end to skip the comment
				yp.move(len(yp.readBytes) - 1)
				break
			}
		}

		if char == tkStringDelim1 || char == tkStringDelim2 {
			// @todo Handle string delimiter escape
			if inString && char == strDelim {
				inString = false
				strDelim = ""
			} else if !inString {
				inString = true
				strDelim = char
			}
		}

		value += char

		if !yp.next() {
			break
		}
	}

	if inString {
		err := yp.err("Syntax Error! Unclosed string")
		return value, err
	}

	return value, nil
}

// read returns the reads "length" characters from the read bytes.
// The length can be negative in order to read backwards.
func (yp *configParser) read(length int) string {
	if length == 0 {
		return ""
	}

	newPos := int(yp.readCursor) + length

	if newPos < 0 || newPos > len(yp.readBytes)-1 {
		return ""
	}

	if length < 0 {
		return string(yp.readBytes[newPos:yp.readCursor])
	}

	return string(yp.readBytes[yp.readCursor:newPos])
}

// move sets the position of the read cursor of "shift" relative steps.
func (yp *configParser) move(shift int) bool {
	pos := int(yp.readCursor) + shift

	if pos < 0 || pos > len(yp.readBytes)-1 {
		return false
	}

	yp.readCursor = uint(pos)

	return true
}

// err returns an error with the given message and additional parser context.
func (yp *configParser) err(msg string) error {
	errorMsg := msg +
		fmt.Sprint(" (l: "+fmt.Sprint(yp.line)+", c: "+fmt.Sprint(yp.readCursor)+")")

	return errors.New(errorMsg)
}
