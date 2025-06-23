import { Skeleton } from "@/components/ui/skeleton"

export function ProductGridSkeleton() {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      {Array.from({ length: 8 }).map((_, index) => (
        <div key={index} className="space-y-3">
          <Skeleton className="aspect-square w-full" />
          <Skeleton className="h-4 w-2/3" />
          <Skeleton className="h-4 w-full" />
          <div className="flex justify-between items-center">
            <Skeleton className="h-4 w-1/3" />
            <Skeleton className="h-4 w-1/4" />
          </div>
          <div className="flex gap-2">
            <Skeleton className="h-9 flex-1" />
            <Skeleton className="h-9 flex-1" />
          </div>
        </div>
      ))}
    </div>
  )
} 