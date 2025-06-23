import { cn } from "@/shared/lib/utils"

interface LoadingProps {
  className?: string
  size?: "sm" | "md" | "lg"
  variant?: "default" | "primary" | "secondary"
}

export function Loading({ className, size = "md", variant = "default" }: LoadingProps) {
  const sizeClasses = {
    sm: "h-4 w-4",
    md: "h-8 w-8",
    lg: "h-12 w-12",
  }

  const variantClasses = {
    default: "border-gray-200",
    primary: "border-primary",
    secondary: "border-secondary",
  }

  return (
    <div className={cn("flex items-center justify-center", className)}>
      <div
        className={cn(
          "animate-spin rounded-full border-4 border-t-transparent",
          sizeClasses[size],
          variantClasses[variant]
        )}
      />
    </div>
  )
} 