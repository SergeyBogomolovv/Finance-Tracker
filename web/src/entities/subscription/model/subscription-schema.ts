import { z } from 'zod'

export const subscriptionSchema = z.object({
  id: z.number(),
  name: z.string(),
  notes: z.string(),
  amount: z.number(),
  currency: z.enum(['RUB', 'USD']),
  service: z.enum([
    'yandex_plus',
    'sberprime',
    'spotify',
    'icloud',
    'vpn',
    'bank',
    'transport',
    'software',
    'shopping',
    'education',
    'music',
    'other',
  ]),
  next_payment_date: z.number(),
  is_autopay: z.boolean(),
  status: z.enum(['active', 'cancelled', 'trial', 'expired']),
  frequency: z.enum(['year', 'half_year', 'quarter', 'month', 'week', 'once']),
  notify: z.boolean(),
  link: z.string().optional(),
})

export type Subscription = z.infer<typeof subscriptionSchema>
