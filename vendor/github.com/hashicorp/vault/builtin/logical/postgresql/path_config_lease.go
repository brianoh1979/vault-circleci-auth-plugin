package postgresql

import (
	"fmt"
	"time"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathConfigLease(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "config/lease",
		Fields: map[string]*framework.FieldSchema{
			"lease": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Default lease for roles.",
			},

			"lease_max": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Maximum time a credential is valid for.",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathLeaseRead,
			logical.UpdateOperation: b.pathLeaseWrite,
		},

		HelpSynopsis:    pathConfigLeaseHelpSyn,
		HelpDescription: pathConfigLeaseHelpDesc,
	}
}

func (b *backend) pathLeaseWrite(
	req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	leaseRaw := d.Get("lease").(string)
	leaseMaxRaw := d.Get("lease_max").(string)

	lease, err := time.ParseDuration(leaseRaw)
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf(
			"Invalid lease: %s", err)), nil
	}
	leaseMax, err := time.ParseDuration(leaseMaxRaw)
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf(
			"Invalid lease: %s", err)), nil
	}

	// Store it
	entry, err := logical.StorageEntryJSON("config/lease", &configLease{
		Lease:    lease,
		LeaseMax: leaseMax,
	})
	if err != nil {
		return nil, err
	}
	if err := req.Storage.Put(entry); err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathLeaseRead(
	req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	lease, err := b.Lease(req.Storage)

	if err != nil {
		return nil, err
	}
	if lease == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"lease":     lease.Lease.String(),
			"lease_max": lease.LeaseMax.String(),
		},
	}, nil
}

type configLease struct {
	Lease    time.Duration
	LeaseMax time.Duration
}

const pathConfigLeaseHelpSyn = `
Configure the default lease information for generated credentials.
`

const pathConfigLeaseHelpDesc = `
This configures the default lease information used for credentials
generated by this backend. The lease specifies the duration that a
credential will be valid for, as well as the maximum session for
a set of credentials.

The format for the lease is "1h" or integer and then unit. The longest
unit is hour.
`