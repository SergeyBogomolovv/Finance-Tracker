import { z } from 'zod'

export const profileSchema = z.object({
  avatar_id: z.string().optional(),
  email: z.string().email(),
  full_name: z.string().optional(),
  provider: z.string(),
  user_id: z.number(),
})

export type Profile = z.infer<typeof profileSchema>
