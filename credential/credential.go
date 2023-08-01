package credential

import (
	"context"

	"github.com/intelops/go-common/credentials"
	"github.com/pkg/errors"
)

const (
	credentialType = "cluster-cred"
)

func GetGenericCredential(ctx context.Context, Entity, CredIdentifier string) (map[string]string, error) {
	credReader, err := credentials.NewCredentialReader(ctx)
	if err != nil {
		err = errors.WithMessage(err, "error in initializing credential reader")
		return nil, err
	}
	cred, err := credReader.GetCredential(context.Background(), credentialType, Entity, CredIdentifier)
	if err != nil {
		err = errors.WithMessage(err, "error in reading credential")
		return nil, err
	}

	return cred, nil
}
