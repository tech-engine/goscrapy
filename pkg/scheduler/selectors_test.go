package scheduler

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectors(t *testing.T) {
	html := `
	<html>
		<body>
			<div id="main" class="content">
				<h1>Title</h1>
				<p class="intro">Introduction paragraph 1</p>
				<a href="http://example.com">Example Link</a>
				<p>This is test paragraph</p>
				<p class="intro" data-mg="test">Introduction paragraph 3</p>
			</div>
		</body>
	</html>
	`

	selector, err := NewSelector(strings.NewReader(html))

	assert.NoError(t, err)

	cssSelector := selector.Css("p.intro")

	cssNodes := cssSelector.GetAll()
	assert.Len(t, cssNodes, 2, "expected nodes=2, got=%s", len(cssNodes))

	cssNodesTexts := cssSelector.Text()
	assert.Equal(t, "Introduction paragraph 1", cssNodesTexts[0], "expected paragraph text=Introduction paragraph 1, got=%s", cssNodesTexts[0])

	xpathSelector := selector.Xpath("//p[@data-mg='test']")

	xpathNodes := xpathSelector.GetAll()
	assert.Len(t, xpathNodes, 1, "expected xpath nodes=1, got=%s", len(xpathNodes))

	xpathNodesTexts := xpathSelector.Text()
	assert.Len(t, xpathNodesTexts, 1, "expected xpathNodesTexts=1, got=%s", len(xpathNodesTexts))
	assert.Equal(t, "Introduction paragraph 3", xpathNodesTexts[0], "expected paragraph text=Introduction paragraph 3, got=%s", xpathNodesTexts[0])

	attrValues := selector.Css("a").Attr("href")
	assert.Len(t, xpathNodesTexts, 1, "expected attrValues=1, got=%s", len(attrValues))
	assert.Equal(t, "http://example.com", attrValues[0], "expected href=http://example.com, got=%s", attrValues[0])

	noCssElements := selector.Css("p.box").GetAll()
	assert.Empty(t, noCssElements, "expected element=0, got=%s", len(noCssElements))

	noXpathElements := selector.Xpath("//p[@class='test']").GetAll()
	assert.Empty(t, noXpathElements, "expected element=0, got=%s", len(noXpathElements))

}
