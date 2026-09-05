import { UserAccountsPage } from './UserAccountsPage';
import { clientUserAccountsApi } from '../../api/users';

export function ClientUsersPage() {
  return (
    <UserAccountsPage
      api={clientUserAccountsApi}
      title="Client Users"
      permissionPrefix="client_user"
      fixedRole="client"
    />
  );
}
