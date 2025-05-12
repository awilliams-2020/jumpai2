import { Navigate } from 'react-router-dom';

interface ProtectedRouteProps {
  children: React.ReactNode;
}

export default function ProtectedRoute({ children }: ProtectedRouteProps) {
  const token = localStorage.getItem('token');
  const userProfile = localStorage.getItem('userProfile');

  if (!token || !userProfile) {
    // Redirect to login if no token or profile
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
} 