package app

const toursSummaryHTML = `<!doctype html>
<html lang="en">
    <head>
        <meta charset="UTF-8" />
        <title>Tours Availability Summary</title>
        <link
            rel="stylesheet"
            href="https://cdn.jsdelivr.net/npm/uikit@3.23.7/dist/css/uikit.min.css"
        />

        <script src="https://cdn.jsdelivr.net/npm/uikit@3.23.7/dist/js/uikit.min.js"></script>
        <script src="https://cdn.jsdelivr.net/npm/uikit@3.23.7/dist/js/uikit-icons.min.js"></script>
    </head>
    <body>
    	{{ range . -}}
        <div class="uk-container uk-margin-top uk-margin-bottom">
            <div
                class="uk-card uk-card-default uk-card-body uk-margin-auto"
            >
                <h3 class="uk-card-title"><a href={{ .Url }}>{{ .Name }}</a></h3>
                <p class="uk-text-meta">{{ .Uuid }}</p>
                <ul class="uk-list uk-list-divider">
                    <li>
                        <strong>Latest Tour Date:</strong> {{ .AvailabilityDate.Format "Mon, 02 Jan 2006 15:04:05 MST" }}
                    </li>
                    <li><strong>Recorded At:</strong> {{ .RecordedAt.Format "Mon, 02 Jan 2006 15:04:05 MST" }}</li>
                </ul>
            </div>
        </div>
        {{ end }}
    </body>
</html>`
