package ldap

import (
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathLogin(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: `login`,
		Fields: map[string]*framework.FieldSchema{
			"username": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "DN (distinguished name) to be used for login.",
			},

			"password": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Password for this user.",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.WriteOperation: b.pathLogin,
		},

		HelpSynopsis:    pathLoginSyn,
		HelpDescription: pathLoginDesc,
	}
}

func (b *backend) pathLogin(
	req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	policies, resp, err := b.Login(req, username, password)
	if len(policies) == 0 {
		return resp, err
	}

	sort.Strings(policies)

	return &logical.Response{
		Auth: &logical.Auth{
			Policies: policies,
			Metadata: map[string]string{
				"username": username,
				"password": password,
				"policies": strings.Join(policies, ","),
			},
			DisplayName: username,
		},
	}, nil
}

func (b *backend) pathLoginRenew(
	req *logical.Request, d *framework.FieldData) (*logical.Response, error) {

	username := req.Auth.Metadata["username"]
	password := req.Auth.Metadata["password"]
	prevpolicies := req.Auth.Metadata["password"]

	policies, resp, err := b.Login(req, username, password)
	if len(policies) == 0 {
		return resp, err
	}

	sort.Strings(policies)
	if strings.Join(policies, ",") != prevpolicies {
		return logical.ErrorResponse("policies have changed, revoking login"), nil
	}

	return framework.LeaseExtend(1*time.Hour, 0)(req, d)
}

const pathLoginSyn = `
Log in with a username and password.
`

const pathLoginDesc = `
This endpoint authenticates using a username and password.
`
