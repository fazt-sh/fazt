import { Card, CardBody } from '../ui';
import { Skeleton } from '../ui';

export function LoginSkeleton() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-[rgb(var(--bg-base))] px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          {/* Logo/Title skeleton */}
          <div className="flex justify-center items-center gap-2 mb-3">
            <Skeleton variant="rect" width={120} height={48} className="rounded-lg" />
          </div>
          <Skeleton variant="text" width={200} height={20} className="mx-auto" />
        </div>

        {/* Login form card skeleton */}
        <Card variant="bordered" className="overflow-hidden">
          <CardBody className="p-8">
            <div className="space-y-5">
              {/* Username input skeleton */}
              <div>
                <Skeleton variant="text" width={80} height={20} className="mb-2" />
                <Skeleton variant="rect" height={48} className="rounded-lg" />
              </div>

              {/* Password input skeleton */}
              <div>
                <Skeleton variant="text" width={80} height={20} className="mb-2" />
                <Skeleton variant="rect" height={48} className="rounded-lg" />
              </div>

              {/* Submit button skeleton */}
              <Skeleton variant="rect" height={48} className="rounded-lg mt-6" />
            </div>
          </CardBody>
        </Card>

        {/* Footer text skeleton */}
        <div className="mt-6">
          <Skeleton variant="text" width={150} height={16} className="mx-auto" />
        </div>
      </div>
    </div>
  );
}