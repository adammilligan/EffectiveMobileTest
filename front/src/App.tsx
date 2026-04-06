import {
  Alert,
  Box,
  Button,
  Chip,
  Container,
  CssBaseline,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from '@mui/material'
import axios from 'axios'
import { useEffect, useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'

type Subscription = {
  id: string
  service_name: string
  price: number
  user_id: string
  start_date: string
  end_date?: string | null
  created_at: string
  updated_at: string
}

type CreateSubscriptionRequest = {
  service_name: string
  price: number
  user_id: string
  start_date: string
  end_date?: string
}

const monthRegex = /^(0[1-9]|1[0-2])-[0-9]{4}$/

const createSchema = z
  .object({
    service_name: z.string().trim().min(1, 'Укажи название подписки'),
    price: z.coerce.number().refine((v) => Number.isFinite(v), 'Цена должна быть числом').int('Цена должна быть целым числом').min(0, 'Цена должна быть >= 0'),
    user_id: z.string().uuid('Некорректный UUID'),
    start_date: z.string().regex(monthRegex, 'Формат даты: MM-YYYY'),
    end_date: z
      .string()
      .trim()
      .optional()
      .refine((v) => v == null || v === '' || monthRegex.test(v), 'Формат даты: MM-YYYY'),
  })
  .superRefine((v, ctx) => {
    if (v.end_date == null || v.end_date === '') {
      return
    }
    const [sm, sy] = v.start_date.split('-')
    const [em, ey] = v.end_date.split('-')
    const startIndex = Number(sy) * 12 + Number(sm)
    const endIndex = Number(ey) * 12 + Number(em)
    if (!Number.isFinite(startIndex) || !Number.isFinite(endIndex)) {
      return
    }
    if (endIndex < startIndex) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['end_date'],
        message: 'Дата окончания должна быть не раньше даты начала',
      })
    }
  })

type CreateFormValues = z.input<typeof createSchema>
type CreateFormSubmitValues = z.output<typeof createSchema>

export default function App() {
  const [subscriptions, setSubscriptions] = useState<Subscription[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isSubmitting },
    reset,
  } = useForm<CreateFormValues, undefined, CreateFormSubmitValues>({
    resolver: zodResolver(createSchema),
    defaultValues: {
      service_name: 'Yandex Plus',
      price: '400',
      user_id: '60601fee-2bf1-4721-ae6f-7636e79a0cba',
      start_date: '07-2025',
      end_date: '',
    },
    mode: 'onBlur',
  })

  const api = useMemo(() => axios.create({ baseURL: '/api' }), [])

  const watchedUserId = watch('user_id')

  const userLabels = useMemo(() => {
    const ids: string[] = []
    for (const s of subscriptions) {
      if (!ids.includes(s.user_id)) {
        ids.push(s.user_id)
      }
    }
    return new Map(ids.map((id, idx) => [id, `Пользователь ${idx + 1}`]))
  }, [subscriptions])

  const load = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const res = await api.get<Subscription[]>('/subscriptions', {
        params: { user_id: watchedUserId || undefined, limit: 50, offset: 0 },
      })
      setSubscriptions(res.data)
    } catch {
      setError('Не удалось загрузить подписки')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    void load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const onSubmit = handleSubmit(async (values) => {
    setError(null)

    const payload: CreateSubscriptionRequest = {
      service_name: values.service_name,
      price: values.price,
      user_id: values.user_id,
      start_date: values.start_date,
    }
    if (values.end_date != null && values.end_date.trim() !== '') {
      payload.end_date = values.end_date.trim()
    }

    try {
      await api.post('/subscriptions', payload)
      await load()
      reset({ ...values, end_date: '' })
    } catch {
      setError('Не удалось создать подписку (проверь поля)')
    }
  })

  return (
    <>
      <CssBaseline />
      <Container maxWidth="lg" sx={{ py: 4 }}>
        <Stack spacing={3}>
          <Typography variant="h4">Подписки</Typography>

          <Paper variant="outlined" sx={{ p: 2 }}>
            <Typography variant="h6" sx={{ mb: 2 }}>
              Добавить подписку
            </Typography>

            <Stack spacing={2} component="form" onSubmit={(e) => void onSubmit(e)}>
              <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
                <TextField
                  label="Service name"
                  {...register('service_name')}
                  error={errors.service_name != null}
                  helperText={errors.service_name?.message}
                  fullWidth
                />
                <TextField
                  label="Price (rub)"
                  {...register('price')}
                  error={errors.price != null}
                  helperText={errors.price?.message}
                  fullWidth
                />
              </Stack>
              <TextField
                label="User UUID"
                {...register('user_id')}
                error={errors.user_id != null}
                helperText={errors.user_id?.message}
                fullWidth
              />
              <Stack direction={{ xs: 'column', md: 'row' }} spacing={2}>
                <TextField
                  label="Start date (MM-YYYY)"
                  {...register('start_date')}
                  error={errors.start_date != null}
                  helperText={errors.start_date?.message}
                  fullWidth
                />
                <TextField
                  label="End date (MM-YYYY, optional)"
                  {...register('end_date')}
                  error={errors.end_date != null}
                  helperText={errors.end_date?.message}
                  fullWidth
                />
              </Stack>

              <Box>
                <Button variant="contained" type="submit" disabled={isSubmitting}>
                  Создать
                </Button>
                <Button sx={{ ml: 2 }} onClick={() => void load()} disabled={isLoading}>
                  Обновить
                </Button>
              </Box>
            </Stack>

            {error ? (
              <Alert severity="error" sx={{ mt: 2 }}>
                {error}
              </Alert>
            ) : null}
          </Paper>

          <Paper variant="outlined" sx={{ p: 2 }}>
            <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
              <Typography variant="h6">Список</Typography>
              <Typography variant="body2" color="text.secondary">
                {isLoading ? 'Загрузка…' : `Записей: ${subscriptions.length}`}
              </Typography>
            </Stack>

            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Пользователь</TableCell>
                  <TableCell>Подписка</TableCell>
                  <TableCell align="right">Цена, ₽/мес</TableCell>
                  <TableCell>Начало</TableCell>
                  <TableCell>Окончание</TableCell>
                  <TableCell>Статус</TableCell>
                  <TableCell>Создано</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {subscriptions.map((s) => {
                  const userLabel = userLabels.get(s.user_id) ?? 'Пользователь'
                  const isActive = s.end_date == null
                  return (
                    <TableRow key={s.id}>
                      <TableCell>{userLabel}</TableCell>
                      <TableCell>{s.service_name}</TableCell>
                      <TableCell align="right">{s.price}</TableCell>
                      <TableCell>{s.start_date}</TableCell>
                      <TableCell>{s.end_date ?? '—'}</TableCell>
                      <TableCell>
                        <Chip
                          label={isActive ? 'OK' : 'Закрыта'}
                          size="small"
                          color={isActive ? 'success' : 'default'}
                          variant={isActive ? 'filled' : 'outlined'}
                        />
                      </TableCell>
                      <TableCell>
                        {new Date(s.created_at).toLocaleString()}
                      </TableCell>
                    </TableRow>
                  )
                })}

                {subscriptions.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7}>
                      <Box sx={{ py: 2, color: 'text.secondary' }}>
                        Пока нет подписок
                      </Box>
                    </TableCell>
                  </TableRow>
                ) : null}
              </TableBody>
            </Table>
          </Paper>
        </Stack>
      </Container>
    </>
  )
}
