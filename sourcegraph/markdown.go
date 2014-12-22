package sourcegraph

import "sourcegraph.com/sourcegraph/go-sourcegraph/router"

type MarkdownService interface {
	Render(markdown []byte, opt MarkdownOpt) ([]byte, Response, error)
}

type markdownService struct {
	client *Client
}

type MarkdownRequestBody struct {
	Markdown []byte
	MarkdownOpt
}

type MarkdownOpt struct {
	EnableCheckboxes bool
}

func (m *markdownService) Render(markdown []byte, opt MarkdownOpt) ([]byte, Response, error) {
	url, err := m.client.URL(router.Markdown, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	req, err := m.client.NewRequest("POST", url.String(), &MarkdownRequestBody{
		Markdown:    markdown,
		MarkdownOpt: opt,
	})
	if err != nil {
		return nil, nil, err
	}

	var rendered []byte
	resp, err := m.client.Do(req, &rendered)
	if err != nil {
		return nil, resp, err
	}

	return rendered, resp, nil
}
