import { Card, CardBody } from '../ui';
import { Skeleton } from '../ui';

export function DashboardSkeleton() {
  return (
    <div>
      {/* Page Header skeleton */}
      <div className="flex items-start justify-between mb-8">
        <div>
          <Skeleton variant="rect" width={200} height={40} className="mb-2" />
          <Skeleton variant="text" width={300} height={20} />
        </div>
        <Skeleton variant="rect" width={140} height={40} className="rounded-lg" />
      </div>

      {/* Stats Grid skeleton */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        {Array.from({ length: 4 }).map((_, i) => (
          <Card key={i} variant="bordered" className="hover-lift">
            <CardBody className="p-6">
              <div className="flex items-start justify-between mb-4">
                <Skeleton variant="circle" width={44} height={44} />
                <Skeleton variant="text" width={60} height={20} className="ml-auto" />
              </div>
              <div className="space-y-1">
                <Skeleton variant="text" width={50} height={16} />
                <Skeleton variant="rect" width={100} height={36} />
              </div>
            </CardBody>
          </Card>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Quick Actions skeleton */}
        <Card variant="bordered" className="p-6">
          <Skeleton variant="rect" width={180} height={28} className="mb-4" />
          <div className="flex flex-wrap gap-3">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} variant="rect" width={100} height={40} className="rounded-lg" />
            ))}
          </div>
        </Card>

        {/* Recent Activity Terminal skeleton */}
        <Card variant="bordered" className="p-0 overflow-hidden">
          {/* Terminal header */}
          <div className="terminal-header">
            <div className="flex items-center gap-2">
              <div className="terminal-dot red"></div>
              <div className="terminal-dot yellow"></div>
              <div className="terminal-dot green"></div>
            </div>
            <div className="flex-1 text-center">
              <Skeleton variant="text" width={120} height={16} className="mx-auto" />
            </div>
          </div>
          <div className="p-4 font-mono text-sm space-y-1">
            {Array.from({ length: 7 }).map((_, i) => (
              <Skeleton key={i} variant="text" width={i % 2 === 0 ? '80%' : '100%'} height={20} />
            ))}
          </div>
        </Card>
      </div>
    </div>
  );
}