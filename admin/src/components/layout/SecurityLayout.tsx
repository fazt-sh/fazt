import { Outlet } from 'react-router-dom';
import { Breadcrumbs } from '../ui/Breadcrumbs';

export function SecurityLayout() {
  return (
    <div>
      <Breadcrumbs />
      <Outlet />
    </div>
  );
}
