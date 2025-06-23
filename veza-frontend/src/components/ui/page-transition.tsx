import { motion } from "framer-motion"
import { cn } from "@/shared/lib/utils"

interface PageTransitionProps extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode
}

export function PageTransition({
  children,
  className,
  ...props
}: PageTransitionProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -20 }}
      transition={{
        type: "spring",
        stiffness: 260,
        damping: 20,
      }}
      className={cn("min-h-screen", className)}
      {...props}
    >
      {children}
    </motion.div>
  )
} 