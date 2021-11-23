// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package catprivilege

import (
	"testing"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
)

const nonSystemDatabaseID = 51

func TestGrantDefaultPrivileges(t *testing.T) {
	defer leaktest.AfterTest(t)()

	fooUser := security.MakeSQLUsernameFromPreNormalizedString("foo")
	barUser := security.MakeSQLUsernameFromPreNormalizedString("bar")
	bazUser := security.MakeSQLUsernameFromPreNormalizedString("baz")
	creatorUser := security.MakeSQLUsernameFromPreNormalizedString("creator")

	testCases := []struct {
		defaultPrivilegesRole descpb.DefaultPrivilegesRole
		privileges            privilege.List
		grantees              []security.SQLUsername
		targetObject          tree.AlterDefaultPrivilegesTargetObject
		objectCreator         security.SQLUsername
	}{
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{Role: creatorUser},
			privileges:            privilege.List{privilege.ALL},
			grantees:              []security.SQLUsername{fooUser},
			targetObject:          tree.Tables,
			objectCreator:         creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{Role: creatorUser},
			privileges:            privilege.List{privilege.ALL},
			grantees:              []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:          tree.Tables,
			objectCreator:         creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{Role: creatorUser},
			privileges:            privilege.List{privilege.ALL},
			grantees:              []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:          tree.Sequences,
			objectCreator:         creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{Role: creatorUser},
			privileges:            privilege.List{privilege.ALL},
			grantees:              []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:          tree.Types,
			objectCreator:         creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{Role: creatorUser},
			privileges:            privilege.List{privilege.ALL},
			grantees:              []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:          tree.Schemas,
			objectCreator:         creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{Role: creatorUser},
			privileges:            privilege.List{privilege.SELECT, privilege.DELETE},
			grantees:              []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:          tree.Tables,
			objectCreator:         creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{Role: creatorUser},
			privileges:            privilege.List{privilege.SELECT, privilege.DELETE},
			grantees:              []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:          tree.Sequences,
			objectCreator:         creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{Role: creatorUser},
			privileges:            privilege.List{privilege.USAGE},
			grantees:              []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:          tree.Types,
			objectCreator:         creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{Role: creatorUser},
			privileges:            privilege.List{privilege.USAGE},
			grantees:              []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:          tree.Schemas,
			objectCreator:         creatorUser,
		},
		/* Test cases for ForAllRoles */
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{ForAllRoles: true},
			privileges:            privilege.List{privilege.ALL},
			grantees:              []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:          tree.Tables,
			objectCreator:         creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{ForAllRoles: true},
			privileges:            privilege.List{privilege.ALL},
			grantees:              []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:          tree.Sequences,
			objectCreator:         creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{ForAllRoles: true},
			privileges:            privilege.List{privilege.ALL},
			grantees:              []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:          tree.Types,
			objectCreator:         creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{ForAllRoles: true},
			privileges:            privilege.List{privilege.ALL},
			grantees:              []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:          tree.Schemas,
			objectCreator:         creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{ForAllRoles: true},
			privileges:            privilege.List{privilege.SELECT, privilege.DELETE},
			grantees:              []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:          tree.Tables,
			objectCreator:         creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{ForAllRoles: true},
			privileges:            privilege.List{privilege.SELECT, privilege.DELETE},
			grantees:              []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:          tree.Sequences,
			objectCreator:         creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{ForAllRoles: true},
			privileges:            privilege.List{privilege.USAGE},
			grantees:              []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:          tree.Types,
			objectCreator:         creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{ForAllRoles: true},
			privileges:            privilege.List{privilege.USAGE},
			grantees:              []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:          tree.Schemas,
			objectCreator:         creatorUser,
		},
	}

	for _, tc := range testCases {
		defaultPrivilegeDescriptor := MakeNewDefaultPrivilegeDescriptor()
		defaultPrivileges := NewMutableDefaultPrivileges(defaultPrivilegeDescriptor)

		defaultPrivileges.GrantDefaultPrivileges(tc.defaultPrivilegesRole, tc.privileges, tc.grantees, tc.targetObject)

		newPrivileges := defaultPrivileges.CreatePrivilegesFromDefaultPrivileges(
			nonSystemDatabaseID, tc.objectCreator, tc.targetObject, &descpb.PrivilegeDescriptor{},
		)

		for _, grantee := range tc.grantees {
			for _, privilege := range tc.privileges {
				if !newPrivileges.CheckPrivilege(grantee, privilege) {
					t.Errorf("expected %s to have %s privilege", grantee, privilege)
				}
			}
		}
	}
}

func TestRevokeDefaultPrivileges(t *testing.T) {
	defer leaktest.AfterTest(t)()

	fooUser := security.MakeSQLUsernameFromPreNormalizedString("foo")
	barUser := security.MakeSQLUsernameFromPreNormalizedString("bar")
	bazUser := security.MakeSQLUsernameFromPreNormalizedString("baz")
	creatorUser := security.MakeSQLUsernameFromPreNormalizedString("creator")

	testCases := []struct {
		defaultPrivilegesRole                                 descpb.DefaultPrivilegesRole
		grantPrivileges, revokePrivileges, expectedPrivileges privilege.List
		grantees                                              []security.SQLUsername
		targetObject                                          tree.AlterDefaultPrivilegesTargetObject
		objectCreator                                         security.SQLUsername
	}{
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{Role: creatorUser},
			grantPrivileges:       privilege.List{privilege.ALL},
			revokePrivileges:      privilege.List{privilege.SELECT},
			expectedPrivileges: privilege.List{
				privilege.CREATE, privilege.DROP, privilege.GRANT, privilege.INSERT,
				privilege.DELETE, privilege.UPDATE, privilege.ZONECONFIG,
			},
			grantees:      []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:  tree.Tables,
			objectCreator: creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{Role: creatorUser},
			grantPrivileges:       privilege.List{privilege.ALL},
			revokePrivileges:      privilege.List{privilege.SELECT},
			expectedPrivileges: privilege.List{
				privilege.CREATE, privilege.DROP, privilege.GRANT, privilege.INSERT,
				privilege.DELETE, privilege.UPDATE, privilege.ZONECONFIG,
			},
			grantees:      []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:  tree.Sequences,
			objectCreator: creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{Role: creatorUser},
			grantPrivileges:       privilege.List{privilege.ALL},
			revokePrivileges:      privilege.List{privilege.USAGE},
			expectedPrivileges: privilege.List{
				privilege.GRANT,
			},
			grantees:      []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:  tree.Types,
			objectCreator: creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{Role: creatorUser},
			grantPrivileges:       privilege.List{privilege.ALL},
			revokePrivileges:      privilege.List{privilege.USAGE},
			expectedPrivileges: privilege.List{
				privilege.CREATE, privilege.GRANT,
			},
			grantees:      []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:  tree.Schemas,
			objectCreator: creatorUser,
		},
		/* Test cases for ForAllRoles */
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{ForAllRoles: true},
			grantPrivileges:       privilege.List{privilege.ALL},
			revokePrivileges:      privilege.List{privilege.SELECT},
			expectedPrivileges: privilege.List{
				privilege.CREATE, privilege.DROP, privilege.GRANT, privilege.INSERT,
				privilege.DELETE, privilege.UPDATE, privilege.ZONECONFIG,
			},
			grantees:      []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:  tree.Sequences,
			objectCreator: creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{ForAllRoles: true},
			grantPrivileges:       privilege.List{privilege.ALL},
			revokePrivileges:      privilege.List{privilege.USAGE},
			expectedPrivileges: privilege.List{
				privilege.GRANT,
			},
			grantees:      []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:  tree.Types,
			objectCreator: creatorUser,
		},
		{
			defaultPrivilegesRole: descpb.DefaultPrivilegesRole{ForAllRoles: true},
			grantPrivileges:       privilege.List{privilege.ALL},
			revokePrivileges:      privilege.List{privilege.USAGE},
			expectedPrivileges: privilege.List{
				privilege.CREATE, privilege.GRANT,
			},
			grantees:      []security.SQLUsername{fooUser, barUser, bazUser},
			targetObject:  tree.Schemas,
			objectCreator: creatorUser,
		},
	}

	for _, tc := range testCases {
		defaultPrivilegeDescriptor := MakeNewDefaultPrivilegeDescriptor()
		defaultPrivileges := NewMutableDefaultPrivileges(defaultPrivilegeDescriptor)

		defaultPrivileges.GrantDefaultPrivileges(tc.defaultPrivilegesRole, tc.grantPrivileges, tc.grantees, tc.targetObject)
		defaultPrivileges.RevokeDefaultPrivileges(tc.defaultPrivilegesRole, tc.revokePrivileges, tc.grantees, tc.targetObject)

		newPrivileges := defaultPrivileges.CreatePrivilegesFromDefaultPrivileges(
			nonSystemDatabaseID, tc.objectCreator, tc.targetObject, &descpb.PrivilegeDescriptor{},
		)

		for _, grantee := range tc.grantees {
			for _, privilege := range tc.expectedPrivileges {
				if !newPrivileges.CheckPrivilege(grantee, privilege) {
					t.Errorf("expected %s to have %s privilege", grantee, privilege)
				}
			}
		}
	}
}

func TestRevokeDefaultPrivilegesFromEmptyList(t *testing.T) {
	defer leaktest.AfterTest(t)()

	defaultPrivilegeDescriptor := MakeNewDefaultPrivilegeDescriptor()
	defaultPrivileges := NewMutableDefaultPrivileges(defaultPrivilegeDescriptor)
	creatorUser := security.MakeSQLUsernameFromPreNormalizedString("creator")
	fooUser := security.MakeSQLUsernameFromPreNormalizedString("foo")
	defaultPrivileges.RevokeDefaultPrivileges(descpb.DefaultPrivilegesRole{
		Role: creatorUser,
	}, privilege.List{privilege.ALL}, []security.SQLUsername{fooUser}, tree.Tables)

	newPrivileges := defaultPrivileges.CreatePrivilegesFromDefaultPrivileges(
		nonSystemDatabaseID, creatorUser, tree.Tables, &descpb.PrivilegeDescriptor{},
	)

	if newPrivileges.AnyPrivilege(fooUser) {
		t.Errorf("expected %s to not have any privileges", fooUser)
	}
}

func TestCreatePrivilegesFromDefaultPrivilegesForSystemDatabase(t *testing.T) {
	defer leaktest.AfterTest(t)()

	defaultPrivilegeDescriptor := MakeNewDefaultPrivilegeDescriptor()
	defaultPrivileges := NewMutableDefaultPrivileges(defaultPrivilegeDescriptor)
	creatorUser := security.MakeSQLUsernameFromPreNormalizedString("creator")
	newPrivileges := defaultPrivileges.CreatePrivilegesFromDefaultPrivileges(
		keys.SystemDatabaseID, creatorUser, tree.Tables, &descpb.PrivilegeDescriptor{},
	)

	if !newPrivileges.Owner().IsNodeUser() {
		t.Errorf("expected owner to be node, owner was %s", newPrivileges.Owner())
	}
}

func TestDefaultDefaultPrivileges(t *testing.T) {
	defer leaktest.AfterTest(t)()

	defaultPrivilegeDescriptor := MakeNewDefaultPrivilegeDescriptor()
	defaultPrivileges := NewMutableDefaultPrivileges(defaultPrivilegeDescriptor)
	creatorUser := security.MakeSQLUsernameFromPreNormalizedString("creator")

	targetObjectTypes := tree.GetAlterDefaultPrivilegesTargetObjects()
	for _, targetObject := range targetObjectTypes {
		newPrivileges := defaultPrivileges.CreatePrivilegesFromDefaultPrivileges(
			nonSystemDatabaseID, creatorUser, targetObject, &descpb.PrivilegeDescriptor{},
		)

		if !newPrivileges.CheckPrivilege(creatorUser, privilege.ALL) {
			t.Errorf("expected creator to have ALL privileges on %s", targetObject)
		}

		if targetObject == tree.Types {
			if !newPrivileges.CheckPrivilege(security.PublicRoleName(), privilege.USAGE) {
				t.Errorf("expected %s to have %s on types", security.PublicRoleName(), privilege.USAGE)
			}
		}
	}
}

func TestDefaultPrivileges(t *testing.T) {
	defer leaktest.AfterTest(t)()

	// The ID chosen doesn't matter as long as it's not the system db ID.
	defaultDatabaseID := descpb.ID(50)

	type userAndGrants struct {
		user   security.SQLUsername
		grants privilege.List
	}
	testCases := []struct {
		objectCreator          security.SQLUsername
		defaultPrivilegesRole  security.SQLUsername
		dbID                   descpb.ID
		targetObject           tree.AlterDefaultPrivilegesTargetObject
		userAndGrants          []userAndGrants
		expectedGrantsOnObject []userAndGrants
	}{
		{
			// Altering default privileges on the system database normally wouldn't
			// be possible but we do it here via directly altering the default
			// privilege descriptor here.
			// The purpose of this test however is to show that even after altering
			// the default privileges, if we create an object in the system database,
			// the only privileges on the object are ALL privileges for root and
			// admin.
			objectCreator:         security.MakeSQLUsernameFromPreNormalizedString("creator"),
			defaultPrivilegesRole: security.MakeSQLUsernameFromPreNormalizedString("creator"),
			targetObject:          tree.Tables,
			dbID:                  keys.SystemDatabaseID,
			userAndGrants: []userAndGrants{
				{
					user:   security.MakeSQLUsernameFromPreNormalizedString("foo"),
					grants: privilege.List{privilege.SELECT},
				},
			},
			expectedGrantsOnObject: []userAndGrants{
				{
					user:   security.RootUserName(),
					grants: privilege.List{privilege.ALL},
				},
				{
					user:   security.AdminRoleName(),
					grants: privilege.List{privilege.ALL},
				},
			},
		},
		{
			objectCreator:         security.MakeSQLUsernameFromPreNormalizedString("creator"),
			defaultPrivilegesRole: security.MakeSQLUsernameFromPreNormalizedString("creator"),
			targetObject:          tree.Tables,
			dbID:                  defaultDatabaseID,
			userAndGrants: []userAndGrants{
				{
					user:   security.MakeSQLUsernameFromPreNormalizedString("foo"),
					grants: privilege.List{privilege.SELECT},
				},
			},
			expectedGrantsOnObject: []userAndGrants{
				{
					user:   security.MakeSQLUsernameFromPreNormalizedString("foo"),
					grants: privilege.List{privilege.SELECT},
				},
			},
		},
		{
			objectCreator:         security.MakeSQLUsernameFromPreNormalizedString("creator"),
			defaultPrivilegesRole: security.MakeSQLUsernameFromPreNormalizedString("creator"),
			targetObject:          tree.Tables,
			dbID:                  defaultDatabaseID,
			userAndGrants: []userAndGrants{
				{
					user:   security.MakeSQLUsernameFromPreNormalizedString("foo"),
					grants: privilege.List{privilege.ALL},
				},
			},
			expectedGrantsOnObject: []userAndGrants{
				{
					user:   security.MakeSQLUsernameFromPreNormalizedString("foo"),
					grants: privilege.List{privilege.ALL},
				},
			},
		},
		{
			objectCreator:         security.MakeSQLUsernameFromPreNormalizedString("creator"),
			defaultPrivilegesRole: security.MakeSQLUsernameFromPreNormalizedString("creator"),
			targetObject:          tree.Tables,
			dbID:                  defaultDatabaseID,
			userAndGrants: []userAndGrants{
				{
					user:   security.MakeSQLUsernameFromPreNormalizedString("foo"),
					grants: privilege.List{privilege.SELECT},
				},
			},
			expectedGrantsOnObject: []userAndGrants{
				{
					user:   security.MakeSQLUsernameFromPreNormalizedString("foo"),
					grants: privilege.List{privilege.SELECT},
				},
			},
		},
		{
			objectCreator:         security.MakeSQLUsernameFromPreNormalizedString("creator"),
			defaultPrivilegesRole: security.MakeSQLUsernameFromPreNormalizedString("creator"),
			targetObject:          tree.Tables,
			dbID:                  defaultDatabaseID,
			userAndGrants: []userAndGrants{
				{
					user:   security.MakeSQLUsernameFromPreNormalizedString("foo"),
					grants: privilege.List{privilege.ALL},
				},
			},
			expectedGrantsOnObject: []userAndGrants{
				{
					user:   security.MakeSQLUsernameFromPreNormalizedString("foo"),
					grants: privilege.List{privilege.ALL},
				},
			},
		},
		{
			objectCreator:         security.MakeSQLUsernameFromPreNormalizedString("creator"),
			defaultPrivilegesRole: security.MakeSQLUsernameFromPreNormalizedString("creator"),
			targetObject:          tree.Tables,
			dbID:                  defaultDatabaseID,
			userAndGrants: []userAndGrants{
				{
					user:   security.MakeSQLUsernameFromPreNormalizedString("foo"),
					grants: privilege.List{privilege.SELECT},
				},
				{
					user:   security.MakeSQLUsernameFromPreNormalizedString("bar"),
					grants: privilege.List{privilege.ALL},
				},
			},
			expectedGrantsOnObject: []userAndGrants{
				{
					user:   security.MakeSQLUsernameFromPreNormalizedString("foo"),
					grants: privilege.List{privilege.SELECT},
				},
				{
					user:   security.MakeSQLUsernameFromPreNormalizedString("bar"),
					grants: privilege.List{privilege.ALL},
				},
			},
		},
		{
			// In this case, we ALTER DEFAULT PRIVILEGES for the role foo.
			// However the default privileges are retrieved for bar, thus
			// we don't expect any privileges on the object.
			objectCreator:         security.MakeSQLUsernameFromPreNormalizedString("foo"),
			defaultPrivilegesRole: security.MakeSQLUsernameFromPreNormalizedString("bar"),
			targetObject:          tree.Tables,
			dbID:                  defaultDatabaseID,
			userAndGrants: []userAndGrants{
				{
					user:   security.MakeSQLUsernameFromPreNormalizedString("foo"),
					grants: privilege.List{privilege.SELECT},
				},
				{
					user:   security.MakeSQLUsernameFromPreNormalizedString("bar"),
					grants: privilege.List{privilege.ALL},
				},
			},
			expectedGrantsOnObject: []userAndGrants{},
		},
	}
	for _, tc := range testCases {
		defaultPrivilegeDescriptor := MakeNewDefaultPrivilegeDescriptor()
		defaultPrivileges := NewMutableDefaultPrivileges(defaultPrivilegeDescriptor)

		for _, userAndGrant := range tc.userAndGrants {
			defaultPrivileges.GrantDefaultPrivileges(
				descpb.DefaultPrivilegesRole{Role: tc.defaultPrivilegesRole},
				userAndGrant.grants,
				[]security.SQLUsername{userAndGrant.user},
				tc.targetObject,
			)
		}

		createdPrivileges := defaultPrivileges.CreatePrivilegesFromDefaultPrivileges(
			tc.dbID,
			tc.objectCreator,
			tc.targetObject,
			&descpb.PrivilegeDescriptor{},
		)

		for _, userAndGrant := range tc.expectedGrantsOnObject {
			for _, grant := range userAndGrant.grants {
				if !createdPrivileges.CheckPrivilege(userAndGrant.user, grant) {
					t.Errorf("expected to find %s privilege for %s", grant.String(), userAndGrant.user)
				}
			}
		}

	}
}

func TestModifyDefaultDefaultPrivileges(t *testing.T) {
	defer leaktest.AfterTest(t)()

	testCases := []struct {
		targetObject             tree.AlterDefaultPrivilegesTargetObject
		revokeAndGrantPrivileges privilege.List
	}{
		{
			targetObject:             tree.Tables,
			revokeAndGrantPrivileges: privilege.List{privilege.SELECT},
		},
		{
			targetObject:             tree.Sequences,
			revokeAndGrantPrivileges: privilege.List{privilege.SELECT},
		},
		{
			targetObject:             tree.Types,
			revokeAndGrantPrivileges: privilege.List{privilege.USAGE},
		},
		{
			targetObject:             tree.Schemas,
			revokeAndGrantPrivileges: privilege.List{privilege.USAGE},
		},
	}

	for _, tc := range testCases {
		defaultPrivilegeDescriptor := MakeNewDefaultPrivilegeDescriptor()
		defaultPrivileges := NewMutableDefaultPrivileges(defaultPrivilegeDescriptor)
		creatorUser := security.MakeSQLUsernameFromPreNormalizedString("creator")

		defaultPrivilegesForCreator := defaultPrivileges.defaultPrivilegeDescriptor.
			FindOrCreateUser(descpb.DefaultPrivilegesRole{
				Role: creatorUser,
			})

		defaultPrivileges.RevokeDefaultPrivileges(
			descpb.DefaultPrivilegesRole{Role: creatorUser},
			tc.revokeAndGrantPrivileges,
			[]security.SQLUsername{creatorUser},
			tc.targetObject,
		)
		if GetRoleHasAllPrivilegesOnTargetObject(defaultPrivilegesForCreator, tc.targetObject) {
			t.Errorf("expected role to not have ALL privileges on %s", tc.targetObject)
		}
		defaultPrivileges.GrantDefaultPrivileges(
			descpb.DefaultPrivilegesRole{Role: creatorUser},
			tc.revokeAndGrantPrivileges,
			[]security.SQLUsername{creatorUser},
			tc.targetObject,
		)
		if !GetRoleHasAllPrivilegesOnTargetObject(defaultPrivilegesForCreator, tc.targetObject) {
			t.Errorf("expected role to have ALL privileges on %s", tc.targetObject)
		}
	}
}

func TestModifyDefaultDefaultPrivilegesForPublic(t *testing.T) {
	defer leaktest.AfterTest(t)()

	defaultPrivilegeDescriptor := MakeNewDefaultPrivilegeDescriptor()
	defaultPrivileges := NewMutableDefaultPrivileges(defaultPrivilegeDescriptor)
	creatorUser := security.MakeSQLUsernameFromPreNormalizedString("creator")

	defaultPrivilegesForCreator := defaultPrivileges.defaultPrivilegeDescriptor.
		FindOrCreateUser(descpb.DefaultPrivilegesRole{
			Role: creatorUser,
		})

	defaultPrivileges.RevokeDefaultPrivileges(
		descpb.DefaultPrivilegesRole{Role: creatorUser},
		privilege.List{privilege.USAGE},
		[]security.SQLUsername{security.PublicRoleName()},
		tree.Types,
	)
	if GetPublicHasUsageOnTypes(defaultPrivilegesForCreator) {
		t.Errorf("expected public to not have USAGE privilege on types")
	}
	defaultPrivileges.GrantDefaultPrivileges(
		descpb.DefaultPrivilegesRole{Role: creatorUser},
		privilege.List{privilege.USAGE},
		[]security.SQLUsername{security.PublicRoleName()},
		tree.Types,
	)
	if !GetPublicHasUsageOnTypes(defaultPrivilegesForCreator) {
		t.Errorf("expected public to have USAGE privilege on types")
	}
}
