{{$status := getStatus .Result -}}
{{$description := getDescription .Result -}}
{{$query := getQuery .Args -}}
{{$action := getAction .Args -}}

{{if eq $status "offering"}}

#### 🌐 Web Search Available

{{$description}}
{{if .Result.context_size}}

- **Context Size:** {{.Result.context_size}}
  {{end}}
  {{if .Result.user_location}}
- **Search Location:** {{range $key, $value := .Result.user_location}}{{$key}}: {{$value}} {{end}}
  {{end}}
- **Status:** ✅ Success

{{else if eq $status "used"}}

#### 🔍 Web Search Executed

{{$description}}

**Query:** `{{$query}}`
{{if .Result.call_id}}**Call ID:** {{.Result.call_id}}{{end}}

- **Status:** ✅ Success

{{else if eq $status "completed"}}

#### ✅ Web Search Session Complete

{{$description}}

{{if .Result.query_count}}**Queries Performed:** {{.Result.query_count}}{{end}}
{{if .Result.citation_count}}**Citations Found:** {{.Result.citation_count}}{{end}}

- **Status:** ✅ Success

{{if .Result.queries}}
**Search Terms:**
{{range .Result.queries}}

- `{{.}}`
  {{end}}
  {{end}}

{{if .Result.citations}}
**Sources Referenced:**
{{range .Result.citations}}

- [{{.}}]({{.}})
  {{end}}
  {{end}}

{{else if eq $status "api_error"}}

#### ❌ Web Search Error

{{$description}}
{{if .Result.error}}
**Error:** {{.Result.error}}
{{end}}

- **Status:** ❌ Failed

{{else}}

#### 🌐 Web Search Activity

{{$description}}

{{if $query}}**Query:** `{{$query}}`{{end}}
{{if .Result.data}}
**Data:** {{.Result.data}}
{{end}}

- **Status:** ✅ Success
  {{end}}

---
