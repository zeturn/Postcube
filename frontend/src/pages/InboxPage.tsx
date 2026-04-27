import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  deleteInboxQuestion,
  fetchInbox,
  fetchMyBox,
  updateInboxQuestion,
  updateMyBox,
  type MyBoxResponse,
  type Question,
} from '../api/client'
import { useAuth } from '../hooks/useAuth'

const presetColors = ['#fff4d6', '#ffe2ec', '#e4f8e9', '#e5f2ff', '#f3e8ff', '#f5f1e8']

function formatDateTime(input: string) {
  return new Date(input).toLocaleString()
}

export default function InboxPage() {
  const { user, logout } = useAuth()

  const [box, setBox] = useState<MyBoxResponse | null>(null)
  const [questions, setQuestions] = useState<Question[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [draftTitle, setDraftTitle] = useState('')
  const [draftAnswers, setDraftAnswers] = useState<Record<number, string>>({})
  const [draftColors, setDraftColors] = useState<Record<number, string>>({})
  const [savingQuestionId, setSavingQuestionId] = useState<number | null>(null)
  const [savingTitle, setSavingTitle] = useState(false)

  const stats = useMemo(() => box?.stats, [box])

  useEffect(() => {
    const load = async () => {
      setLoading(true)
      setError(null)
      try {
        const [inboxData, myBoxData] = await Promise.all([fetchInbox(), fetchMyBox()])
        setQuestions(inboxData)
        setBox(myBoxData)
      } catch (err) {
        setError((err as Error).message)
      } finally {
        setLoading(false)
      }
    }

    load()
  }, [])

  useEffect(() => {
    const answerMap: Record<number, string> = {}
    const colorMap: Record<number, string> = {}

    for (const question of questions) {
      answerMap[question.id] = question.answer || ''
      colorMap[question.id] = question.background_color || '#fff4d6'
    }

    setDraftAnswers(answerMap)
    setDraftColors(colorMap)
  }, [questions])

  useEffect(() => {
    setDraftTitle(box?.user.box_title || '')
  }, [box])

  const handleSaveTitle = async () => {
    if (!draftTitle.trim()) {
      setError('Box title cannot be empty.')
      return
    }

    setSavingTitle(true)
    setError(null)
    try {
      const updated = await updateMyBox({ box_title: draftTitle.trim() })
      setBox((prev) => {
        if (!prev) return prev
        return { ...prev, user: updated }
      })
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setSavingTitle(false)
    }
  }

  const handleSaveQuestion = async (questionId: number) => {
    setSavingQuestionId(questionId)
    setError(null)
    try {
      const updated = await updateInboxQuestion(questionId, {
        answer: draftAnswers[questionId] ?? '',
        background_color: draftColors[questionId],
      })

      setQuestions((prev) => prev.map((q) => (q.id === questionId ? updated : q)))
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setSavingQuestionId(null)
    }
  }

  const handleDeleteQuestion = async (questionId: number) => {
    const sure = window.confirm('Delete this question permanently?')
    if (!sure) return

    setError(null)
    try {
      await deleteInboxQuestion(questionId)
      setQuestions((prev) => prev.filter((q) => q.id !== questionId))
      setBox((prev) => {
        if (!prev) return prev
        const answered = prev.stats.answered - (questions.find((q) => q.id === questionId)?.status === 'answered' ? 1 : 0)
        const total = Math.max(prev.stats.total - 1, 0)
        return {
          ...prev,
          stats: {
            total,
            answered: Math.max(answered, 0),
            unanswered: Math.max(total - Math.max(answered, 0), 0),
          },
        }
      })
    } catch (err) {
      setError((err as Error).message)
    }
  }

  const publicBoxURL = box ? `${window.location.origin}/u/${box.user.slug}` : ''

  return (
    <div className="mx-auto max-w-6xl px-4 py-6 md:px-6 md:py-10">
      <header className="mb-6 rounded-3xl bg-gradient-to-r from-ink-800 via-slate-800 to-sky-700 p-6 text-white card-shadow">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <p className="text-xs font-semibold tracking-[0.2em] text-ink-200">POSTCUBE INBOX</p>
            <h1 className="mt-2 text-3xl font-bold">Welcome, {user?.name}</h1>
            <p className="mt-2 text-sm text-ink-100">Answer anonymous questions, tune card backgrounds, and curate your public feed.</p>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            {box && (
              <Link to={`/u/${box.user.slug}`} className="rounded-xl border border-white/40 px-4 py-2 text-sm font-semibold text-white">
                Open Public Box
              </Link>
            )}
            <button
              onClick={logout}
              className="cursor-pointer rounded-xl bg-white px-4 py-2 text-sm font-semibold text-ink-800 transition hover:bg-ink-100"
            >
              Logout
            </button>
          </div>
        </div>
      </header>

      {error && <div className="mb-4 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>}

      {loading && (
        <div className="rounded-2xl border border-ink-100 bg-white p-6 text-sm text-ink-500 card-shadow">Loading inbox...</div>
      )}

      {!loading && box && (
        <>
          <section className="mb-6 grid gap-4 md:grid-cols-4">
            <div className="rounded-2xl border border-ink-100 bg-white p-4 card-shadow md:col-span-2">
              <label className="text-xs font-semibold tracking-wide text-ink-500">Box title</label>
              <div className="mt-2 flex flex-col gap-3 sm:flex-row">
                <input
                  value={draftTitle}
                  onChange={(e) => setDraftTitle(e.target.value)}
                  className="h-11 w-full rounded-xl border border-ink-200 px-3 text-sm outline-none focus:border-brand-500"
                  maxLength={120}
                />
                <button
                  onClick={handleSaveTitle}
                  disabled={savingTitle}
                  className="h-11 cursor-pointer rounded-xl bg-brand-500 px-4 text-sm font-semibold text-white transition hover:bg-brand-600 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {savingTitle ? 'Saving...' : 'Save'}
                </button>
              </div>
            </div>

            <div className="rounded-2xl border border-ink-100 bg-white p-4 card-shadow">
              <p className="text-sm text-ink-500">Total</p>
              <p className="mt-1 text-3xl font-bold text-ink-800">{stats?.total ?? 0}</p>
            </div>

            <div className="rounded-2xl border border-ink-100 bg-white p-4 card-shadow">
              <p className="text-sm text-ink-500">Answered</p>
              <p className="mt-1 text-3xl font-bold text-emerald-700">{stats?.answered ?? 0}</p>
            </div>
          </section>

          <section className="mb-6 rounded-2xl border border-ink-100 bg-white p-4 card-shadow">
            <p className="text-xs font-semibold tracking-wide text-ink-500">Public URL</p>
            <div className="mt-2 flex flex-col gap-3 sm:flex-row">
              <input readOnly value={publicBoxURL} className="h-11 w-full rounded-xl border border-ink-200 bg-ink-50 px-3 text-sm text-ink-700" />
              <button
                onClick={async () => {
                  await navigator.clipboard.writeText(publicBoxURL)
                }}
                className="h-11 cursor-pointer rounded-xl border border-ink-300 px-4 text-sm font-semibold text-ink-700 transition hover:bg-ink-100"
              >
                Copy Link
              </button>
            </div>
          </section>

          <section className="space-y-4">
            {questions.map((question) => (
              <article key={question.id} className="rounded-2xl border border-ink-100 bg-white p-5 card-shadow">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <p className="text-sm font-semibold text-ink-800">{question.anonymous_name}</p>
                  <div className="flex items-center gap-2 text-xs">
                    <span
                      className={`rounded-full px-2 py-1 font-semibold ${
                        question.status === 'answered' ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-700'
                      }`}
                    >
                      {question.status === 'answered' ? 'Answered' : 'Pending'}
                    </span>
                    <span className="text-ink-400">{formatDateTime(question.created_at)}</span>
                  </div>
                </div>

                <div className="mt-3 rounded-xl border border-ink-100 bg-ink-50 p-3 text-sm text-ink-700">{question.content}</div>

                <div className="mt-4">
                  <label className="text-xs font-semibold tracking-wide text-ink-500">Answer</label>
                  <textarea
                    value={draftAnswers[question.id] ?? ''}
                    onChange={(e) =>
                      setDraftAnswers((prev) => ({
                        ...prev,
                        [question.id]: e.target.value,
                      }))
                    }
                    className="mt-2 h-28 w-full resize-none rounded-xl border border-ink-200 px-3 py-2 text-sm outline-none focus:border-brand-500"
                    placeholder="Leave empty and save if you want this to stay pending."
                  />
                </div>

                <div className="mt-4 flex flex-wrap items-center gap-2">
                  {presetColors.map((color) => (
                    <button
                      key={`${question.id}-${color}`}
                      type="button"
                      onClick={() =>
                        setDraftColors((prev) => ({
                          ...prev,
                          [question.id]: color,
                        }))
                      }
                      className="h-8 w-8 cursor-pointer rounded-full border border-ink-200"
                      style={{ backgroundColor: color }}
                      aria-label={`Set color ${color}`}
                    />
                  ))}
                  <input
                    type="color"
                    value={draftColors[question.id] ?? '#fff4d6'}
                    onChange={(e) =>
                      setDraftColors((prev) => ({
                        ...prev,
                        [question.id]: e.target.value,
                      }))
                    }
                    className="h-8 w-10 cursor-pointer rounded border border-ink-200 bg-white"
                  />
                </div>

                <div className="mt-4 flex flex-wrap gap-3">
                  <button
                    onClick={() => handleSaveQuestion(question.id)}
                    disabled={savingQuestionId === question.id}
                    className="h-10 cursor-pointer rounded-xl bg-ink-800 px-4 text-sm font-semibold text-white transition hover:bg-ink-700 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {savingQuestionId === question.id ? 'Saving...' : 'Save Answer + Color'}
                  </button>
                  <button
                    onClick={() => handleDeleteQuestion(question.id)}
                    className="h-10 cursor-pointer rounded-xl border border-red-200 px-4 text-sm font-semibold text-red-600 transition hover:bg-red-50"
                  >
                    Delete
                  </button>
                </div>
              </article>
            ))}

            {questions.length === 0 && (
              <div className="rounded-2xl border border-ink-100 bg-white p-8 text-center text-sm text-ink-500 card-shadow">
                Your inbox is empty. Share your public link to start receiving questions.
              </div>
            )}
          </section>
        </>
      )}
    </div>
  )
}
