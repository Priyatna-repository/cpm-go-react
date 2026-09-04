import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';

export function RequirePermission({ permission }: { permission: string }) {
  const { can } = useAuth();

  if (!can(permission)) {
    return <Navigate to="/dashboard" replace />;
  }

  return <Outlet />;
}
