import { UserAccountsPage } from './UserAccountsPage';
import { internalUsersApi } from '../../api/users';

const ROLE_OPTIONS = [
  { value: 'manager', label: 'Manager' },
  { value: 'team member', label: 'Team Member' },
];

export function UsersPage() {
  return (
    <UserAccountsPage
      api={internalUsersApi}
      title="Users"
      permissionPrefix="user"
      roleOptions={ROLE_OPTIONS}
    />
  );
}
