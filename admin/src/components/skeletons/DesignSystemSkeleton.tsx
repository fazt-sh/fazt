import { Card, CardBody, CardHeader } from '../ui';
import { Skeleton } from '../ui';

export function DesignSystemSkeleton() {
  return (
    <div className="p-6">
      {/* Page Header skeleton */}
      <div className="flex items-start justify-between mb-8">
        <div>
          <Skeleton variant="rect" width={300} height={40} className="mb-2" />
          <Skeleton variant="text" width={400} height={20} />
        </div>
        <Skeleton variant="rect" width={160} height={40} className="rounded-lg" />
      </div>

      {/* Buttons Section skeleton */}
      <Card className="mb-6">
        <CardHeader>
          <Skeleton variant="rect" width={150} height={28} />
        </CardHeader>
        <CardBody>
          <div className="space-y-4">
            <div className="flex flex-wrap gap-3">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} variant="rect" width={100} height={40} className="rounded-lg" />
              ))}
            </div>
            <div className="flex items-center gap-3">
              <Skeleton variant="text" width={60} height={20} />
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} variant="rect" width={80} height={36} className="rounded-lg" />
              ))}
            </div>
          </div>
        </CardBody>
      </Card>

      {/* Inputs Section skeleton */}
      <Card className="mb-6">
        <CardHeader>
          <Skeleton variant="rect" width={100} height={28} />
        </CardHeader>
        <CardBody>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i}>
                <Skeleton variant="text" width={100} height={20} className="mb-2" />
                <Skeleton variant="rect" height={48} className="rounded-lg" />
              </div>
            ))}
          </div>
          <div className="mt-4">
            <Skeleton variant="text" width={140} height={20} className="mb-2" />
            <Skeleton variant="rect" height={48} className="rounded-lg" />
          </div>
        </CardBody>
      </Card>

      {/* Cards Section skeleton */}
      <Card className="mb-6">
        <CardHeader>
          <Skeleton variant="rect" width={80} height={28} />
        </CardHeader>
        <CardBody>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {Array.from({ length: 3 }).map((_, i) => (
              <Card key={i} variant="bordered" hover>
                <CardBody>
                  <Skeleton variant="rect" width={120} height={24} className="mb-2" />
                  <Skeleton variant="text" lines={2} />
                </CardBody>
              </Card>
            ))}
          </div>
        </CardBody>
      </Card>

      {/* Badges Section skeleton */}
      <Card className="mb-6">
        <CardHeader>
          <Skeleton variant="rect" width={80} height={28} />
        </CardHeader>
        <CardBody>
          <div className="space-y-4">
            {Array.from({ length: 3 }).map((_, groupIndex) => (
              <div key={groupIndex}>
                <Skeleton variant="text" width={100} height={20} className="mb-2" />
                <div className="flex flex-wrap gap-2">
                  {Array.from({ length: 5 }).map((_, i) => (
                    <Skeleton key={i} variant="rect" width={60} height={28} className="rounded-full" />
                  ))}
                </div>
              </div>
            ))}
          </div>
        </CardBody>
      </Card>

      {/* Loading States Section skeleton */}
      <Card className="mb-6">
        <CardHeader>
          <Skeleton variant="rect" width={140} height={28} />
        </CardHeader>
        <CardBody>
          <div className="space-y-4">
            <div>
              <Skeleton variant="text" width={80} height={20} className="mb-2" />
              <div className="flex items-center gap-4">
                {Array.from({ length: 5 }).map((_, i) => (
                  <Skeleton key={i} variant="circle" width={20 + i * 8} height={20 + i * 8} />
                ))}
              </div>
            </div>
            <div>
              <Skeleton variant="text" width={80} height={20} className="mb-2" />
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Skeleton variant="card" />
                <Skeleton variant="card" />
              </div>
            </div>
          </div>
        </CardBody>
      </Card>

      {/* Status Indicators skeleton */}
      <Card>
        <CardHeader>
          <Skeleton variant="rect" width={160} height={28} />
        </CardHeader>
        <CardBody>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className="text-center p-4 rounded-lg border border-[rgb(var(--border-primary))]">
                <Skeleton variant="circle" width={32} height={32} className="mx-auto mb-2" />
                <Skeleton variant="text" width={80} height={20} className="mx-auto" />
              </div>
            ))}
          </div>
        </CardBody>
      </Card>
    </div>
  );
}