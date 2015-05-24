package kingpin

var ManPageTemplate = `{{define "FormatFlags"}}\
{{range .Flags}}\
{{if not .Hidden}}\
.TP
\fB{{if .Short}}-{{.Short|Char}}, {{end}}--{{.Name}}{{if not .IsBoolFlag}}={{.FormatPlaceHolder}}{{end}}\\fR
{{.Help}}
{{end}}\
{{end}}\
{{end}}\

{{define "FormatCommand"}}\
{{if .FlagSummary}} {{.FlagSummary}}{{end}}\
{{range .Args}} {{if not .Required}}[{{end}}<{{.Name}}>{{if .Value|IsCumulative}}...{{end}}{{if not .Required}}]{{end}}{{end}}\
{{end}}\

{{define "FormatCommands"}}\
{{range .FlattenedCommands}}\
.SS
\fB{{.FullCommand}}{{template "FormatCommand" .}}\\fR
.PP
{{.Help}}
{{template "FormatFlags" .}}\
{{end}}\
{{end}}\

{{define "FormatUsage"}}\
{{template "FormatCommand" .}}{{if .Commands}} <command> [<args> ...]{{end}}\\fR
{{end}}\

.TH {{.App.Name}} 1 {{.App.Version}} "{{.App.Author}}"
.SH "NAME"
{{.App.Name}}
.SH "SYNOPSIS"
.TP
\fB{{.App.Name}}{{template "FormatUsage" .App}}
.SH "DESCRIPTION"
{{.App.Help}}
.SH "OPTIONS"
{{template "FormatFlags" .App}}\
{{if .App.Commands}}\
.SH "COMMANDS"
{{template "FormatCommands" .App}}\
{{end}}\
`
