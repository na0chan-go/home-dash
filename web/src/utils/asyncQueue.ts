export type AsyncTask<T> = () => Promise<T>

export function createAsyncQueue(): <T>(task: AsyncTask<T>) => Promise<T> {
  let tail: Promise<void> = Promise.resolve()

  return async function enqueue<T>(task: AsyncTask<T>): Promise<T> {
    const run = tail.then(task)

    // Keep the chain alive even when a task fails.
    tail = run.then(
      () => undefined,
      () => undefined
    )

    return run
  }
}
