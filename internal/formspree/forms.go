package formspree

import "strings"

// Forms is every site's signup form, in site order.
func (p *Project) Forms() ([]Form, error) {
	if err := p.ready(); err != nil {
		return nil, err
	}
	config, err := p.readConfig()
	if err != nil {
		return nil, err
	}

	var forms []Form
	for _, key := range config.formKeys() {
		forms = append(forms, Form{
			Site:     strings.TrimSuffix(key, "-signup"),
			Endpoint: p.endpointFor(key),
		})
	}
	return forms, nil
}
