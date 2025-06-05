package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/shortid"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

var mapper map[string]func(db *gorm.DB, id string) (string, error) = map[string]func(*gorm.DB, string) (string, error){
	"job": func(db *gorm.DB, id string) (string, error) {
		var obj app.RunnerJob
		res := db.First(&obj, "id = ?", id)
		if res.Error != nil {
			return "", errors.Wrap(res.Error, "unable to fetch runner job")
		}

		return obj.OrgID, nil
	},
	"rgr": func(db *gorm.DB, id string) (string, error) {
		var obj app.RunnerGroup
		res := db.First(&obj, "id = ?", id)
		if res.Error != nil {
			return "", errors.Wrap(res.Error, "unable to fetch runner group")
		}

		return obj.OrgID, nil
	},
	"bld": func(db *gorm.DB, id string) (string, error) {
		var obj app.ComponentBuild
		res := db.First(&obj, "id = ?", id)
		if res.Error != nil {
			return "", errors.Wrap(res.Error, "unable to fetch build")
		}

		return obj.OrgID, nil
	},
	"run": func(db *gorm.DB, id string) (string, error) {
		var obj app.Runner
		res := db.First(&obj, "id = ?", id)
		if res.Error != nil {
			return "", errors.Wrap(res.Error, "unable to fetch runner")
		}

		return obj.OrgID, nil
	},
	"org": func(db *gorm.DB, id string) (string, error) {
		var obj app.Org
		res := db.First(&obj, "id = ?", id)
		if res.Error != nil {
			return "", errors.Wrap(res.Error, "unable to fetch org")
		}

		return obj.ID, nil
	},
	"cmp": func(db *gorm.DB, id string) (string, error) {
		var obj app.Component
		res := db.First(&obj, "id = ?", id)
		if res.Error != nil {
			return "", errors.Wrap(res.Error, "unable to fetch component")
		}

		return obj.OrgID, nil
	},
	"app": func(db *gorm.DB, id string) (string, error) {
		var obj app.App
		res := db.First(&obj, "id = ?", id)
		if res.Error != nil {
			return "", errors.Wrap(res.Error, "unable to fetch app")
		}

		return obj.OrgID, nil
	},
	"inl": func(db *gorm.DB, id string) (string, error) {
		var obj app.Install
		res := db.First(&obj, "id = ?", id)
		if res.Error != nil {
			return "", errors.Wrap(res.Error, "unable to fetch install")
		}

		return obj.OrgID, nil
	},
	"dpl": func(db *gorm.DB, id string) (string, error) {
		var obj app.InstallDeploy
		res := db.First(&obj, "id = ?", id)
		if res.Error != nil {
			return "", errors.Wrap(res.Error, "unable to fetch install deploy")
		}

		return obj.OrgID, nil
	},
	"iar": func(db *gorm.DB, id string) (string, error) {
		var obj app.InstallActionWorkflowRun
		res := db.First(&obj, "id = ?", id)
		if res.Error != nil {
			return "", errors.Wrap(res.Error, "unable to fetch install action workflow run")
		}

		return obj.OrgID, nil
	},
	"rop": func(db *gorm.DB, id string) (string, error) {
		var obj app.RunnerOperation
		res := db.First(&obj, "id = ?", id)
		if res.Error != nil {
			return "", errors.Wrap(res.Error, "unable to fetch runner operation")
		}

		return obj.OrgID, nil
	},
	"sbr": func(db *gorm.DB, id string) (string, error) {
		var obj app.InstallSandboxRun
		res := db.First(&obj, "id = ?", id)
		if res.Error != nil {
			return "", errors.Wrap(res.Error, "unable to fetch install sandbox run")
		}

		return obj.OrgID, nil
	},
	"rel": func(db *gorm.DB, id string) (string, error) {
		var obj app.ComponentRelease
		res := db.First(&obj, "id = ?", id)
		if res.Error != nil {
			return "", errors.Wrap(res.Error, "unable to fetch release")
		}

		return obj.OrgID, nil
	},
	"iws": func(db *gorm.DB, id string) (string, error) {
		var obj app.InstallWorkflowStep
		res := db.First(&obj, "id = ?", id)
		if res.Error != nil {
			return "", errors.Wrap(res.Error, "unable to fetch step")
		}

		return obj.OrgID, nil
	},
}

func (m *middleware) setOrgID(ctx *gin.Context) error {
	for _, param := range ctx.Params {
		if !shortid.IsShortID(param.Value) {
			m.l.Debug("not a short id", zap.String("value", param.Value))
			continue
		}

		prefix, err := shortid.GetPrefix(param.Value)
		if err != nil {
			m.l.Debug("no prefix parsed", zap.String("value", param.Value))
			continue
		}

		fn, ok := mapper[prefix]
		if !ok {
			m.l.Error("no prefix mapper found", zap.String("prefix", prefix))
			continue
		}

		db := m.db.WithContext(ctx)
		orgID, err := fn(db, param.Value)
		if err != nil {
			return errors.Wrap(err, "middleware failed to find object")
		}

		cctx.SetOrgIDGinContext(ctx, orgID)
		m.l.Debug("successfully set org id", zap.String("org-id", orgID))
		return nil
	}

	return nil
}
