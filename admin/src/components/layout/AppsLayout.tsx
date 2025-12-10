import { Outlet } from 'react-router-dom';
import { Breadcrumbs } from '../ui/Breadcrumbs';

export function AppsLayout() {
  return (
    <div>
      <Breadcrumbs />
      <Outlet />
    </div>
  );
}
