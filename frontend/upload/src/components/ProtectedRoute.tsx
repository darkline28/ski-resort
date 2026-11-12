import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '@/context/AuthContext';
import { cabinetPath } from '@/utils/routes';
import type { Role } from '@/types';
import { Loader } from './States';

/** Пускает только авторизованных пользователей нужной роли */
export function ProtectedRoute({ roles }: { roles?: Role[] }) {
  const { user, initializing } = useAuth();
  const location = useLocation();

  if (initializing) return <div className="container section"><Loader /></div>;
  if (!user) return <Navigate to="/login" replace state={{ from: location.pathname + location.search }} />;
  if (roles && !roles.includes(user.role)) return <Navigate to={cabinetPath(user)} replace />;
  return <Outlet />;
}
