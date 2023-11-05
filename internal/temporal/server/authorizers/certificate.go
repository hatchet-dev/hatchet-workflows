// Adapted from: https://github.com/hatchet-dev/hatchet/blob/3c2c13168afa1af68d4baaf5ed02c9d49c5f0323/internal/temporal/server/authorizer/authorizer.go

package authorizers

import (
	"context"

	temporal "github.com/hatchet-dev/hatchet-workflows/internal/temporal/server/config"
	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/log"
)

var permittedWorkerAPIs map[string]bool = map[string]bool{
	"/temporal.api.workflowservice.v1.WorkflowService/GetSystemInfo":         true,
	"/temporal.api.workflowservice.v1.WorkflowService/DescribeNamespace":     true,
	"/temporal.api.workflowservice.v1.WorkflowService/PollWorkflowTaskQueue": true,
	"/temporal.api.workflowservice.v1.WorkflowService/PollActivityTaskQueue": true,
}

type CertificateAuthorizer struct {
	config *temporal.Config

	logger log.Logger
}

func NewCertificateAuthorizer(config *temporal.Config, authCfg *config.Authorization, logger log.Logger) *CertificateAuthorizer {
	return &CertificateAuthorizer{config, logger}
}

var (
	decisionAllow = authorization.Result{Decision: authorization.DecisionAllow}
	decisionDeny  = authorization.Result{Decision: authorization.DecisionDeny}
)

// Authorize allows all internal-admin and internal-worker requests. This should not be used in production!
func (a *CertificateAuthorizer) Authorize(_ context.Context, claims *authorization.Claims,
	target *authorization.CallTarget,
) (authorization.Result, error) {
	if claims != nil && claims.Subject == "internal-admin" {
		return decisionAllow, nil
	}

	if claims != nil && claims.Subject == "internal-worker" {
		return decisionAllow, nil
	}

	// Deny any requests that don't match the internal admin user
	return decisionDeny, nil
}

func (a *CertificateAuthorizer) GetClaims(authInfo *authorization.AuthInfo) (*authorization.Claims, error) {
	claims := &authorization.Claims{}

	// if TLS information is passed through, case on trusted CNs
	if authInfo.TLSSubject != nil {
		switch authInfo.TLSSubject.CommonName {
		case "cluster":
			claims.Subject = "internal-admin"
			claims.System = authorization.RoleAdmin
		case "internal-admin":
			claims.Subject = "internal-admin"
			claims.System = authorization.RoleAdmin
		case "internal-worker":
			claims.Subject = "internal-worker"
			claims.System = authorization.RoleWorker
		}
	}

	return claims, nil
}
