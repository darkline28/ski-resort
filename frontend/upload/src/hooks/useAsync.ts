import { useCallback, useEffect, useRef, useState, type DependencyList } from 'react';
import { ApiError } from '@/api/client';

export interface AsyncState<T> {
  data: T | undefined;
  loading: boolean;
  error: ApiError | Error | null;
  reload: () => void;
  setData: (updater: T | ((prev: T | undefined) => T)) => void;
}

/**
 * Загрузка данных с состояниями loading / error / data.
 * Устаревшие ответы (при смене deps) игнорируются.
 */
export function useAsync<T>(fn: () => Promise<T>, deps: DependencyList): AsyncState<T> {
  const [data, setDataState] = useState<T>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<ApiError | Error | null>(null);
  const [tick, setTick] = useState(0);
  const callId = useRef(0);

  useEffect(() => {
    const id = ++callId.current;
    setLoading(true);
    setError(null);
    fn()
      .then((d) => { if (id === callId.current) setDataState(d); })
      .catch((e) => { if (id === callId.current) setError(e); })
      .finally(() => { if (id === callId.current) setLoading(false); });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...deps, tick]);

  const reload = useCallback(() => setTick((t) => t + 1), []);
  const setData = useCallback((u: T | ((prev: T | undefined) => T)) => {
    setDataState((prev) => (typeof u === 'function' ? (u as (p: T | undefined) => T)(prev) : u));
  }, []);

  return { data, loading, error, reload, setData };
}
