package controller

import (
	kutilcore "github.com/appscode/kutil/core/v1"
	kutilrbac "github.com/appscode/kutil/rbac/v1beta1"
	"github.com/k8sdb/apimachinery/apis/kubedb"
	api "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) deleteRole(xdb *api.Xdb) error {
	// Delete existing Roles
	if err := c.Client.RbacV1beta1().Roles(xdb.Namespace).Delete(xdb.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createRole(xdb *api.Xdb) error {
	// Create new Roles
	_, err := kutilrbac.CreateOrPatchRole(
		c.Client,
		metav1.ObjectMeta{
			Name:      xdb.OffshootName(),
			Namespace: xdb.Namespace,
		},
		func(in *rbac.Role) *rbac.Role {
			in.Rules = []rbac.PolicyRule{
				{
					APIGroups:     []string{kubedb.GroupName},
					Resources:     []string{api.ResourceTypeXdb},
					ResourceNames: []string{xdb.Name},
					Verbs:         []string{"get"},
				},
				{
					// TODO. Use this if secret is necessary, Otherwise remove it
					APIGroups:     []string{core.GroupName},
					Resources:     []string{"secrets"},
					ResourceNames: []string{xdb.Spec.DatabaseSecret.SecretName},
					Verbs:         []string{"get"},
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) deleteServiceAccount(xdb *api.Xdb) error {
	// Delete existing ServiceAccount
	if err := c.Client.CoreV1().ServiceAccounts(xdb.Namespace).Delete(xdb.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createServiceAccount(xdb *api.Xdb) error {
	// Create new ServiceAccount
	_, err := kutilcore.CreateOrPatchServiceAccount(
		c.Client,
		metav1.ObjectMeta{
			Name:      xdb.OffshootName(),
			Namespace: xdb.Namespace,
		},
		func(in *core.ServiceAccount) *core.ServiceAccount {
			return in
		},
	)
	return err
}

func (c *Controller) deleteRoleBinding(xdb *api.Xdb) error {
	// Delete existing RoleBindings
	if err := c.Client.RbacV1beta1().RoleBindings(xdb.Namespace).Delete(xdb.OffshootName(), nil); err != nil {
		if !kerr.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *Controller) createRoleBinding(xdb *api.Xdb) error {
	// Ensure new RoleBindings
	_, err := kutilrbac.CreateOrPatchRoleBinding(
		c.Client,
		metav1.ObjectMeta{
			Name:      xdb.OffshootName(),
			Namespace: xdb.Namespace,
		},
		func(in *rbac.RoleBinding) *rbac.RoleBinding {
			in.RoleRef = rbac.RoleRef{
				APIGroup: rbac.GroupName,
				Kind:     "Role",
				Name:     xdb.OffshootName(),
			}
			in.Subjects = []rbac.Subject{
				{
					Kind:      rbac.ServiceAccountKind,
					Name:      xdb.OffshootName(),
					Namespace: xdb.Namespace,
				},
			}
			return in
		},
	)
	return err
}

func (c *Controller) createRBACStuff(xdb *api.Xdb) error {
	// Delete Existing Role
	if err := c.deleteRole(xdb); err != nil {
		return err
	}
	// Create New Role
	if err := c.createRole(xdb); err != nil {
		return err
	}

	// Create New ServiceAccount
	if err := c.createServiceAccount(xdb); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	// Create New RoleBinding
	if err := c.createRoleBinding(xdb); err != nil {
		if !kerr.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

func (c *Controller) deleteRBACStuff(xdb *api.Xdb) error {
	// Delete Existing Role
	if err := c.deleteRole(xdb); err != nil {
		return err
	}

	// Delete ServiceAccount
	if err := c.deleteServiceAccount(xdb); err != nil {
		return err
	}

	// Delete New RoleBinding
	if err := c.deleteRoleBinding(xdb); err != nil {
		return err
	}

	return nil
}
