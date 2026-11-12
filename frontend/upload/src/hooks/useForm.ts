import { useState, type ChangeEvent, type FormEvent } from 'react';
import { ApiError } from '@/api/client';

export type Validator<V> = (values: V) => Partial<Record<keyof V, string>>;

/**
 * Простой хук форм: значения, клиентская валидация, ошибки сервера по полям, состояние отправки.
 */
export function useForm<V extends Record<string, string>>(initial: V, validate: Validator<V>) {
  const [values, setValues] = useState<V>(initial);
  const [errors, setErrors] = useState<Partial<Record<keyof V, string>>>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [touched, setTouched] = useState(false);

  const onChange = (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    const next = { ...values, [name]: value };
    setValues(next);
    if (touched) setErrors(validate(next));
  };

  const handleSubmit = (onValid: (v: V) => Promise<void> | void) => async (e: FormEvent) => {
    e.preventDefault();
    setTouched(true);
    setFormError(null);
    const errs = validate(values);
    setErrors(errs);
    if (Object.values(errs).some(Boolean)) return;
    setSubmitting(true);
    try {
      await onValid(values);
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.fields) setErrors((prev) => ({ ...prev, ...err.fields }));
        setFormError(err.message);
      } else {
        setFormError('Что-то пошло не так. Попробуй ещё раз.');
      }
    } finally {
      setSubmitting(false);
    }
  };

  const reset = (v: V = initial) => { setValues(v); setErrors({}); setTouched(false); setFormError(null); };

  return { values, setValues, errors, formError, submitting, onChange, handleSubmit, reset };
}

// ---- валидаторы ----
export const isEmail = (s: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(s.trim());
export const isPhone = (s: string) => s.replace(/\D/g, '').length >= 10;
