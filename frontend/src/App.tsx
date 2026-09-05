import { Navigate, Route, Routes } from 'react-router-dom';
import { LoginPage } from './pages/Login/LoginPage';
import { DashboardPage } from './pages/Dashboard/DashboardPage';
import { RolesPermissionsPage } from './pages/Settings/RolesPermissionsPage';
import { CompanyPage } from './pages/Settings/CompanyPage';
import { CompaniesPage } from './pages/Clients/CompaniesPage';
import { LabelsPage } from './pages/Settings/LabelsPage';
import { ProtectedRoute } from './routes/ProtectedRoute';
import { RequirePermission } from './routes/RequirePermission';
import { MainLayout } from './layouts/MainLayout';
import { ProjectsPage } from './pages/Projects/ProjectsPage';
import { ClientUsersPage } from './pages/Users/ClientUsersPage';
import { UsersPage } from './pages/Users/UsersPage';

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<MainLayout />}>
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route element={<RequirePermission permission="project.view" />}>
          <Route path="/projects" element={<ProjectsPage />} />
        </Route>
          <Route element={<RequirePermission permission="roles.view" />}>
            <Route path="/settings/roles" element={<RolesPermissionsPage />} />
          </Route>
          <Route element={<RequirePermission permission="owner_company.view" />}>
            <Route path="/settings/company" element={<CompanyPage />} />
          </Route>
          <Route element={<RequirePermission permission="client_company.view" />}>
            <Route path="/clients/companies" element={<CompaniesPage />} />
          </Route>
          <Route element={<RequirePermission permission="labels.view" />}>
            <Route path="/settings/labels" element={<LabelsPage />} />
          </Route>
          <Route element={<RequirePermission permission="user.view" />}>
            <Route path="/settings/users" element={<UsersPage />} />
          </Route>
          <Route element={<RequirePermission permission="client_user.view" />}>
            <Route path="/clients/users" element={<ClientUsersPage />} />
          </Route>
        </Route>
      </Route>
      <Route path="/" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}
