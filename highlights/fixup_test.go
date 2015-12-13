package highlights

import (
	"bytes"
	"strings"
	"testing"
)

func TestFixup(t *testing.T) {
	reader := strings.NewReader(testdocument)
	writer := new(bytes.Buffer)

	ShowDesiredLineInFile("2", reader, writer)
	if a, ea := writer.String(), resultdocument; a != ea {
		t.Errorf("got %v expected %v", a, ea)
	}
}

const testdocument = `<!DOCTYPE html>
<html>
<head>
<meta charset="ISO-8859-1">
<title>GeneratePreviewForURL.m</title>
<link rel="stylesheet" type="text/css" href="highlight.css">
</head>
<body class="hl">
<ol>
<li class="hl"><span class="hl com">/*&nbsp;This&nbsp;code&nbsp;is&nbsp;copyright&nbsp;Nathaniel&nbsp;Gray,&nbsp;licensed&nbsp;under&nbsp;the&nbsp;GPL&nbsp;v2.&nbsp;&nbsp;</span></li>
<li class="hl"><span class="hl com">&nbsp;&nbsp;&nbsp;&nbsp;See&nbsp;LICENSE.txt&nbsp;for&nbsp;details.&nbsp;*/</span></li>
<li class="hl"></li>
</ol>
</body>
</html>
<!--HTML generated by highlight 3.24, http://www.andre-simon.de/-->
`

const resultdocument = `<!DOCTYPE html>
<html>
<head>
<meta charset="ISO-8859-1">
<title>GeneratePreviewForURL.m</title>
<link rel="stylesheet" type="text/css" href="highlight.css">
</head>
<body class="hl">
<ol>
<li class="hl"><span class="hl com">/*&nbsp;This&nbsp;code&nbsp;is&nbsp;copyright&nbsp;Nathaniel&nbsp;Gray,&nbsp;licensed&nbsp;under&nbsp;the&nbsp;GPL&nbsp;v2.&nbsp;&nbsp;</span></li>
<li  id="theline" style="background-color: rgb(80,80,80);" class="hl"><span class="hl com">&nbsp;&nbsp;&nbsp;&nbsp;See&nbsp;LICENSE.txt&nbsp;for&nbsp;details.&nbsp;*/</span></li>
<li class="hl"></li>
</ol>
</body>
</html>
<!--HTML generated by highlight 3.24, http://www.andre-simon.de/-->
<script>
document.getElementById("theline").scrollIntoViewIfNeeded(true);
</script>
`
