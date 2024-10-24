package ssml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSSML(t *testing.T) {

	testStr := `<speak>
					<p>
						<s> This is sentence one. &amp;</s>
						<s> This is sentence two with
							<emphasis level="strong">
								emphasis
							</emphasis> then more text
						</s>
					</p>
				</speak>`

	tokenizer := NewWordTokenizer()
	parser := NewSSMLParser(tokenizer)

	t.Run("Tokenizer", func(t *testing.T) {

		tokens := tokenizer.tokenize(testStr)

		expected := []Token{
			{value: "<speak>"},
			{value: "<p>"},
			{value: "<s>"},
			{value: " This is sentence one. &"},
			{value: "</s>"},
			{value: "<s>"},
			{value: " This is sentence two with"},
			{value: "<emphasis level=\"strong\">"},
			{value: "emphasis"},
			{value: "</emphasis>"},
			{value: " then more text"},
			{value: "</s>"},
			{value: "</p>"},
			{value: "</speak>"},
		}

		assert.Equal(t, expected, tokens)
	})

	t.Run("Tokens", func(t *testing.T) {

		t.Run("isTag", func(t *testing.T) {
			token := &Token{value: "<speak>"}
			assert.Equal(t, true, token.isTag())
			token = &Token{value: "</speak>"}
			assert.Equal(t, true, token.isTag())
			token = &Token{value: "text"}
			assert.Equal(t, false, token.isTag())
		})

		t.Run("isOpeningTag", func(t *testing.T) {
			token := &Token{value: "<speak>"}
			assert.Equal(t, true, token.isOpeningTag())
			token = &Token{value: "</speak>"}
			assert.Equal(t, false, token.isOpeningTag())
			token = &Token{value: "text"}
			assert.Equal(t, false, token.isOpeningTag())
		})

		t.Run("isClosingTag", func(t *testing.T) {
			token := &Token{value: "<speak>"}
			assert.Equal(t, false, token.isClosingTag())
			token = &Token{value: "</speak>"}
			assert.Equal(t, true, token.isClosingTag())
			token = &Token{value: "text"}
			assert.Equal(t, false, token.isClosingTag())
		})
	})

	t.Run("Parser", func(t *testing.T) {

		t.Run("Root tree", func(t *testing.T) {
			root := parser.ParseTree("<speak></speak>")
			expected := &SSMLNode{value: "speak", attributes: []SSMLAttribute{}, children: []*SSMLNode{}}
			assert.Equal(t, expected, root)
		})

		t.Run("Root tree with one child", func(t *testing.T) {
			root := parser.ParseTree("<speak>text</speak>")
			expected := &SSMLNode{
				value:      "speak",
				attributes: []SSMLAttribute{},
				children: []*SSMLNode{{
					value:      "text",
					attributes: []SSMLAttribute{},
					children:   []*SSMLNode{}},
				}}
			assert.Equal(t, expected, root)
		})

		t.Run("Root with children 2 levels", func(t *testing.T) {
			root := parser.ParseTree("<speak>text<emphasis>emp</emphasis></speak>")
			expected := &SSMLNode{
				value:      "speak",
				attributes: []SSMLAttribute{},
				children: []*SSMLNode{{
					value:      "text",
					attributes: []SSMLAttribute{},
					children:   []*SSMLNode{},
				}, {
					value:      "emphasis",
					attributes: []SSMLAttribute{},
					children: []*SSMLNode{{
						value:      "emp",
						attributes: []SSMLAttribute{},
						children:   []*SSMLNode{},
					}},
				}},
			}
			assert.Equal(t, expected, root)
		})

		t.Run("Parse with self closing tags", func(t *testing.T) {
			root := parser.ParseTree("<speak>hello< br/>world!</speak>")
			expected := &SSMLNode{
				value:      "speak",
				attributes: []SSMLAttribute{},
				children: []*SSMLNode{{
					value:      "hello",
					attributes: []SSMLAttribute{},
					children:   []*SSMLNode{},
				}, {
					value:      "br",
					attributes: []SSMLAttribute{},
					children:   []*SSMLNode{},
				}, {
					value:      "world!",
					attributes: []SSMLAttribute{},
					children:   []*SSMLNode{},
				}},
			}
			assert.Equal(t, expected, root)
		})

		t.Run("Parse with attributes", func(t *testing.T) {
			root := parser.ParseTree(`<speak attr="test"></speak>`)
			expected := &SSMLNode{
				value:      "speak",
				attributes: []SSMLAttribute{{name: "attr", value: "test"}},
				children:   []*SSMLNode{},
			}
			assert.Equal(t, expected, root)
		})

		t.Run("Complex Parse with attributes and nested elements", func(t *testing.T) {
			root := parser.ParseTree(`
		<speak version="1.0" xml:lang="en-US">
			<voice name="Matthew" gender="male">
				<s>Here is <emphasis level="strong">important</emphasis> text.</s>
			</voice>
			<voice name="Joanna" gender="female">
				<s>Another sentence with <prosody rate="fast">faster speech</prosody>.</s>
			</voice>
		</speak>`)

			expected := &SSMLNode{
				value: "speak",
				attributes: []SSMLAttribute{
					{name: "version", value: "1.0"},
					{name: "xml:lang", value: "en-US"},
				},
				children: []*SSMLNode{
					{
						value: "voice",
						attributes: []SSMLAttribute{
							{name: "name", value: "Matthew"},
							{name: "gender", value: "male"},
						},
						children: []*SSMLNode{
							{
								value:      "s",
								attributes: []SSMLAttribute{},
								children: []*SSMLNode{
									{
										value:      "Here is ",
										attributes: []SSMLAttribute{},
										children:   []*SSMLNode{},
									},
									{
										value: "emphasis",
										attributes: []SSMLAttribute{
											{name: "level", value: "strong"},
										},
										children: []*SSMLNode{
											{
												value:      "important",
												attributes: []SSMLAttribute{},
												children:   []*SSMLNode{},
											},
										},
									},
									{
										value:      " text.",
										attributes: []SSMLAttribute{},
										children:   []*SSMLNode{},
									},
								},
							},
						},
					},
					{
						value: "voice",
						attributes: []SSMLAttribute{
							{name: "name", value: "Joanna"},
							{name: "gender", value: "female"},
						},
						children: []*SSMLNode{
							{
								value:      "s",
								attributes: []SSMLAttribute{},
								children: []*SSMLNode{
									{
										value:      "Another sentence with ",
										attributes: []SSMLAttribute{},
										children:   []*SSMLNode{},
									},
									{
										value: "prosody",
										attributes: []SSMLAttribute{
											{name: "rate", value: "fast"},
										},
										children: []*SSMLNode{
											{
												value:      "faster speech",
												attributes: []SSMLAttribute{},
												children:   []*SSMLNode{},
											},
										},
									},
									{
										value:      ".",
										attributes: []SSMLAttribute{},
										children:   []*SSMLNode{},
									},
								},
							},
						},
					},
				},
			}

			assert.Equal(t, expected, root)
		})

		t.Run("Root with children 3 levels", func(t *testing.T) {
			root := parser.ParseTree("<speak>text<emphasis>emp</emphasis><emphasis>emp2</emphasis></speak>")
			expected := &SSMLNode{
				value:      "speak",
				attributes: []SSMLAttribute{},
				children: []*SSMLNode{
					{
						value:      "text",
						attributes: []SSMLAttribute{},
						children:   []*SSMLNode{},
					},
					{
						value:      "emphasis",
						attributes: []SSMLAttribute{},
						children: []*SSMLNode{
							{
								value:      "emp",
								attributes: []SSMLAttribute{},
								children:   []*SSMLNode{},
							},
						},
					},
					{
						value:      "emphasis",
						attributes: []SSMLAttribute{},
						children: []*SSMLNode{
							{
								value:      "emp2",
								attributes: []SSMLAttribute{},
								children:   []*SSMLNode{},
							},
						},
					},
				},
			}

			assert.Equal(t, expected, root)
		})

		t.Run("Complex SSML structure with attributes", func(t *testing.T) {
			root := parser.ParseTree(`
		<speak>
			<p>
				<s>Sentence one <emphasis>important</emphasis></s>
				<s>Sentence two with <emphasis>strong emphasis</emphasis></s>
			</p>
		</speak>`)

			expected := &SSMLNode{
				value:      "speak",
				attributes: []SSMLAttribute{},
				children: []*SSMLNode{
					{
						value:      "p",
						attributes: []SSMLAttribute{},
						children: []*SSMLNode{
							{
								value:      "s",
								attributes: []SSMLAttribute{},
								children: []*SSMLNode{
									{
										value:      "Sentence one ",
										attributes: []SSMLAttribute{},
										children:   []*SSMLNode{},
									},
									{
										value:      "emphasis",
										attributes: []SSMLAttribute{},
										children: []*SSMLNode{
											{
												value:      "important",
												attributes: []SSMLAttribute{},
												children:   []*SSMLNode{},
											},
										},
									},
								},
							},
							{
								value:      "s",
								attributes: []SSMLAttribute{},
								children: []*SSMLNode{
									{
										value:      "Sentence two with ",
										attributes: []SSMLAttribute{},
										children:   []*SSMLNode{},
									},
									{
										value:      "emphasis",
										attributes: []SSMLAttribute{},
										children: []*SSMLNode{
											{
												value:      "strong emphasis",
												attributes: []SSMLAttribute{},
												children:   []*SSMLNode{},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			assert.Equal(t, expected, root)
		})

		t.Run("More Complex SSML Structure", func(t *testing.T) {
			root := parser.ParseTree(`
    <speak      >
        <p>
            <s>Introduction sentence with <emphasis>key</emphasis> point</s>
            <s>Another sentence with pause</s>
        </p>
        <p>
            <s>Further details are provided with <emphasis>important</emphasis> remark</s>
            <s>Final thoughts with <prosody>deliberate pacing</prosody></s>
            <s>
                <say-as>SSML</say-as> is complex
				<audio>Audio info not supported</audio>
            </s>
        </p>
    </speak>`)

			expected := &SSMLNode{
				value:      "speak",
				attributes: []SSMLAttribute{},
				children: []*SSMLNode{
					{
						value:      "p",
						attributes: []SSMLAttribute{},
						children: []*SSMLNode{
							{
								value:      "s",
								attributes: []SSMLAttribute{},
								children: []*SSMLNode{
									{
										value:      "Introduction sentence with ",
										attributes: []SSMLAttribute{},
										children:   []*SSMLNode{},
									},
									{
										value:      "emphasis",
										attributes: []SSMLAttribute{},
										children: []*SSMLNode{
											{
												value:      "key",
												attributes: []SSMLAttribute{},
												children:   []*SSMLNode{},
											},
										},
									},
									{
										value:      " point",
										attributes: []SSMLAttribute{},
										children:   []*SSMLNode{},
									},
								},
							},
							{
								value:      "s",
								attributes: []SSMLAttribute{},
								children: []*SSMLNode{
									{
										value:      "Another sentence with pause",
										attributes: []SSMLAttribute{},
										children:   []*SSMLNode{},
									},
								},
							},
						},
					},
					{
						value:      "p",
						attributes: []SSMLAttribute{},
						children: []*SSMLNode{
							{
								value:      "s",
								attributes: []SSMLAttribute{},
								children: []*SSMLNode{
									{
										value:      "Further details are provided with ",
										attributes: []SSMLAttribute{},
										children:   []*SSMLNode{},
									},
									{
										value:      "emphasis",
										attributes: []SSMLAttribute{},
										children: []*SSMLNode{
											{
												value:      "important",
												attributes: []SSMLAttribute{},
												children:   []*SSMLNode{},
											},
										},
									},
									{
										value:      " remark",
										attributes: []SSMLAttribute{},
										children:   []*SSMLNode{},
									},
								},
							},
							{
								value:      "s",
								attributes: []SSMLAttribute{},
								children: []*SSMLNode{
									{
										value:      "Final thoughts with ",
										attributes: []SSMLAttribute{},
										children:   []*SSMLNode{},
									},
									{
										value:      "prosody",
										attributes: []SSMLAttribute{},
										children: []*SSMLNode{
											{
												value:      "deliberate pacing",
												attributes: []SSMLAttribute{},
												children:   []*SSMLNode{},
											},
										},
									},
								},
							},
							{
								value:      "s",
								attributes: []SSMLAttribute{},
								children: []*SSMLNode{
									{
										value:      "say-as",
										attributes: []SSMLAttribute{},
										children: []*SSMLNode{
											{
												value:      "SSML",
												attributes: []SSMLAttribute{},
												children:   []*SSMLNode{},
											},
										},
									},
									{
										value:      " is complex",
										attributes: []SSMLAttribute{},
										children:   []*SSMLNode{},
									},
									{
										value:      "audio",
										attributes: []SSMLAttribute{},
										children: []*SSMLNode{
											{
												value:      "Audio info not supported",
												attributes: []SSMLAttribute{},
												children:   []*SSMLNode{},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			assert.Equal(t, expected, root)
		})
	})
}
