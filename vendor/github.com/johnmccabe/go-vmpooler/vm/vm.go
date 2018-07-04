package vm

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/johnmccabe/go-vmpooler/time"
)

// VM represents a VM known to vmpooler belonging to the authenticated user
type VM struct {
	Hostname string
	Fqdn     string
	Ip       string
	State    string
	Running  float64
	Lifetime float64
	Tags     map[string]string
	Template Template
}

// Template breaks down the vmpooler pool/template into its component parts
type Template struct {
	Id    string
	Os    string
	Osver string
	Arch  string
}

type legacyVM struct {
	Template string
	Lifetime float64
	Running  float64
	State    string
	Tags     map[string]string
	Ip       string
	Domain   string
}

type legacyGetTokenOutput struct {
	User    string
	Created time.PoolerTime
	Last    time.PoolerTime
	Vms     legacyGetTokenVms
}

type legacyGetTokenVms struct {
	Running []string
}

type setLifetimeRequest struct {
	Lifetime int `json:"lifetime"`
}

type setTagsRequest struct {
	Tags map[string]string `json:"tags"`
}

type setResponse struct {
	Ok bool
}

// Get returns details on the specified VM
// Endpoint: GET /v1/vm/<hostname>
func (c *Client) Get(hostname string) (*VM, error) {
	req, err := c.NewRequest("GET", fmt.Sprintf("%s%s/%s", c.Endpoint, "/vm", hostname), nil)
	if err != nil {
		return nil, err
	}

	response := map[string]interface{}{}

	err = c.SendWithAuth(req, &response)
	if err != nil {
		return nil, err
	}

	if err := responseOk(response); err != nil {
		return nil, err
	}

	lvm, ok := response[hostname].(map[string]interface{})
	if !ok {
		return nil, errors.New("unable to parse response")
	}
	vm := &VM{
		Hostname: hostname,
		Fqdn:     fmt.Sprintf("%s.%s", hostname, lvm["domain"].(string)),
		Ip:       lvm["ip"].(string),
		State:    lvm["state"].(string),
		Running:  lvm["running"].(float64),
		Lifetime: lvm["lifetime"].(float64),
		Template: parseTemplate(lvm["template"].(string)),
	}

	if _, ok := lvm["tags"]; ok {
		parsedTags := make(map[string]string)
		for k, v := range lvm["tags"].(map[string]interface{}) {
			parsedTags[k] = v.(string)
		}
		vm.Tags = parsedTags
	}

	return vm, nil
}

// GetAll return all VMs associated with the authenticated user
// Endpoint: GET /v1/vm/<hostname>
func (c *Client) GetAll() ([]VM, error) {
	req, err := c.NewRequest("GET", fmt.Sprintf("%s%s/%s", c.Endpoint, "/token", c.Token), nil)
	if err != nil {
		return nil, err
	}

	response := map[string]interface{}{}

	err = c.SendWithAuth(req, &response)
	if err != nil {
		return nil, err
	}

	if err := responseOk(response); err != nil {
		return nil, err
	}

	lvms := legacyGetTokenOutput{}
	if l, ok := response[c.Token]; !ok {
		return nil, errors.New("unable to parse response")
	} else {
		b, err := json.Marshal(l)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(b, &lvms); err != nil {
			return nil, err
		}
	}

	vms := []VM{}
	if len(lvms.Vms.Running) > 0 {
		for _, hostname := range lvms.Vms.Running {
			vm, err := c.Get(hostname)
			if err != nil {
				return nil, err
			}
			vms = append(vms, *vm)
		}
	}
	return vms, nil
}

// Create returns a VM from the specified pool/template belonging to the authenticated user
// Endpoint: POST /vm/<template>
func (c *Client) Create(template string) (*VM, error) {
	req, err := c.NewRequest("POST", fmt.Sprintf("%s%s/%s", c.Endpoint, "/vm", template), nil)
	if err != nil {
		return nil, err
	}

	response := map[string]interface{}{}

	err = c.SendWithAuth(req, &response)
	if err != nil {
		return nil, err
	}

	if err := responseOk(response); err != nil {
		return nil, err
	}

	lvm, ok := response[template].(map[string]interface{})
	if !ok {
		return nil, errors.New("unable to parse response")
	}

	hostname, ok := lvm["hostname"]
	if !ok {
		return nil, errors.New("unable to find hostname in response")
	}

	return c.Get(hostname.(string))
}

// Delete the the specified VM belonging to the authenticated user
// Endpoint: DELETE /vm/<hostname>
func (c *Client) Delete(hostname string) error {
	req, err := c.NewRequest("DELETE", fmt.Sprintf("%s%s/%s", c.Endpoint, "/vm", hostname), nil)
	if err != nil {
		return err
	}

	response := map[string]interface{}{}

	err = c.SendWithAuth(req, &response)
	if err != nil {
		return err
	}

	if err := responseOk(response); err != nil {
		return err
	}

	return nil
}

// SetLifetime override the lifetime associated with a VM
// Endpoint: PUT /vm/<hostname>
// Payload: {"lifetime":"2"} - lifetime in hours
func (c *Client) SetLifetime(hostname string, hours int) (*VM, error) {
	if hours < 1 {
		return nil, errors.New("lifetime must be greater than 0 hours")
	}
	lifetime := setLifetimeRequest{Lifetime: hours}

	req, err := c.NewRequest("PUT", fmt.Sprintf("%s%s/%s", c.Endpoint, "/vm", hostname), lifetime)
	if err != nil {
		return nil, err
	}

	response := setResponse{}
	err = c.SendWithAuth(req, &response)
	if err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, errors.New("error thrown setting lifetime on VM")
	}

	return c.Get(hostname)
}

// SetTags overrides the tags associated with a VM
// Endpoint: PUT /vm/<hostname>
// Payload: {"tags":{"department":"engineering","user":"jdoe"}}
func (c *Client) SetTags(hostname string, tags map[string]string) (*VM, error) {

	setTags := setTagsRequest{Tags: tags}

	req, err := c.NewRequest("PUT", fmt.Sprintf("%s%s/%s", c.Endpoint, "/vm", hostname), setTags)
	if err != nil {
		return nil, err
	}

	response := setResponse{}
	err = c.SendWithAuth(req, &response)
	if err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, errors.New("error thrown setting tags on VM")
	}

	return c.Get(hostname)
}

// ListTemplates lists all VM templates/pools
// Endpoint: GET /v1/vm
func (c *Client) ListTemplates() ([]string, error) {
	req, err := c.NewRequest("GET", fmt.Sprintf("%s%s", c.Endpoint, "/vm"), nil)
	if err != nil {
		return nil, err
	}

	response := []string{}

	err = c.Send(req, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func parseTemplate(id string) Template {
	parts := strings.Split(id, "-")
	template := Template{
		Id:    id,
		Os:    parts[0],
		Osver: strings.Join(parts[1:len(parts)-1], "-"),
		Arch:  parts[len(parts)-1],
	}
	return template
}

func responseOk(response map[string]interface{}) error {
	if r, ok := response["ok"]; ok {
		if rb, ok := r.(bool); !ok || !rb {
			return errors.New("response from vmpooler not OK")
		}
	} else {
		return errors.New("invalid response returned from vmpooler")
	}
	return nil
}
