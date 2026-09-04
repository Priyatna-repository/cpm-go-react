import { Navigate, Route, Routes } from 'react-router-dom';
import { LoginPage } from './pages/Login/LoginPage';
import { DashboardPage } from './pages/Dashboard/DashboardPage';
import { RolesPermissionsPage } from './pages/Settings/RolesPermissionsPage';
import { CompanyPage } from './pages/Settings/CompanyPage';
import { CompaniesPage } from './pages/Clients/CompaniesPage';
import { ProtectedRoute } from './routes/ProtectedRoute';
import { RequirePermission } from './routes/RequirePermission';

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<ProtectedRoute />}>
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route element={<RequirePermission permission="roles.view" />}>
          <Route path="/settings/roles" element={<RolesPermissionsPage />} />
        </Route>
        <Route element={<RequirePermission permission="owner_company.view" />}>
          <Route path="/settings/company" element={<CompanyPage />} />
        </Route>
        <Route element={<RequirePermission permission="client_company.view" />}>
          <Route path="/clients/companies" element={<CompaniesPage />} />
        </Route>
      </Route>
      <Route path="/" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}
