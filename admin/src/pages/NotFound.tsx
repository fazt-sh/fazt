import { Link } from 'react-router-dom';
import { AlertCircle, Home } from 'lucide-react';

export function NotFound() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh]">
      <AlertCircle className="w-16 h-16 text-tertiary mb-4" />
      <h1 className="text-3xl font-bold text-primary mb-2">404 - Page Not Found</h1>
      <p className="text-secondary mb-6">The page you're looking for doesn't exist.</p>
      <Link
        to="/dashboard"
        className="inline-flex items-center gap-2 px-4 py-2 bg-accent text-white rounded-lg hover:bg-accent/90 transition-colors"
      >
        <Home className="w-4 h-4" />
        Back to Dashboard
      </Link>
    </div>
  );
}
