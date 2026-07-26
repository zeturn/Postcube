import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { fetchPublicBox, submitQuestion, type PublicBoxResponse } from '../api/client'
import { useAuth } from '../hooks/useAuth'
import Button from '../components/ui/Button'
import Input from '../components/ui/Input'
import Textarea from '../components/ui/Textarea'
import Card from '../components/ui/Card'

export default function PublicBoxPage() {
  const { slug = '' } = useParams()
  const { user } = useAuth()

  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [data, setData] = useState<PublicBoxResponse | null>(null)

  const [content, setContent] = useState('')
  const [anonymousName, setAnonymousName] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [submitMessage, setSubmitMessage] = useState<string | null>(null)

  useEffect(() => {
    const load = async () => {
      setLoading(true)
      setError(null)
      try {
        const box = await fetchPublicBox(slug)
        setData(box)
      } catch (err) {
        setError((err as Error).message)
      } finally {
        setLoading(false)
      }
    }

    load()
  }, [slug])

  const totalCount = useMemo(() => {
    if (!data) return 0
    return data.answered.length + data.unanswered.length
  }, [data])

  const onSubmit = async (event: React.FormEvent) => {
    event.preventDefault()
    setSubmitMessage(null)

    if (!content.trim()) {
      setSubmitMessage('Please enter a question first.')
      return
    }

    setSubmitting(true)
    try {
      const created = await submitQuestion(slug, {
        content: content.trim(),
        anonymous_name: anonymousName.trim() || undefined,
      })
      setData((prev) => {
        if (!prev) return prev
        return {
          ...prev,
          unanswered: [created, ...prev.unanswered],
        }
      })
      setContent('')
      setAnonymousName('')
      setSubmitMessage('Your question has been sent.')
    } catch (err) {
      setSubmitMessage((err as Error).message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="mx-auto max-w-5xl px-4 py-6 md:px-6 md:py-10">
      <header className="mb-6 rounded-3xl bg-gradient-to-r from-ink-800 via-ink-700 to-sky-700 p-6 text-white card-shadow">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <p className="text-xs font-semibold tracking-[0.2em] text-ink-200">POSTCUBE</p>
            <h1 className="mt-2 text-3xl font-bold">{data?.owner.box_title || 'Question Box'}</h1>
            <p className="mt-2 text-sm text-ink-100">Ask anything anonymously. Answers and pending messages are visible here.</p>
          </div>
          <div className="flex gap-3">
            {user ? (
              <Link to="/inbox" className="rounded-xl bg-white px-4 py-2 text-sm font-semibold text-ink-800">
                Open Inbox
              </Link>
            ) : (
              <Link to="/login" className="rounded-xl border border-white/40 px-4 py-2 text-sm font-semibold text-white">
                Sign In
              </Link>
            )}
          </div>
        </div>
      </header>

      {loading && (
        <Card className="!p-6 bg-white text-sm text-ink-500">Loading box...</Card>
      )}

      {!loading && error && (
        <Card className="!p-6 border-red-200 bg-red-50 text-sm text-red-700">{error}</Card>
      )}

      {!loading && data && (
        <div className="grid gap-6 lg:grid-cols-[minmax(0,360px)_minmax(0,1fr)]">
          <section className="space-y-4">
            <Card className="!p-5 bg-white">
              <h2 className="text-lg font-bold text-ink-800">Ask anonymously</h2>
              <form className="mt-4 space-y-3" onSubmit={onSubmit}>
                <Input
                  value={anonymousName}
                  onChange={(e) => setAnonymousName(e.target.value)}
                  placeholder="Display name (optional)"
                  maxLength={24}
                />
                <Textarea
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  rows={6}
                  placeholder="Type your anonymous question..."
                  maxLength={500}
                />
                <Button variant="brand" fullWidth type="submit" disabled={submitting}>
                  {submitting ? 'Sending...' : 'Send Question'}
                </Button>
              </form>
              {submitMessage && <p className="mt-3 text-sm text-ink-500">{submitMessage}</p>}
            </Card>

            <Card className="!p-5 bg-white">
              <p className="text-sm text-ink-500">Total posts</p>
              <p className="mt-1 text-3xl font-bold text-ink-800">{totalCount}</p>
              <div className="mt-4 flex gap-2 text-xs">
                <span className="rounded-full bg-emerald-100 px-3 py-1 text-emerald-700">Answered: {data.answered.length}</span>
                <span className="rounded-full bg-amber-100 px-3 py-1 text-amber-700">Pending: {data.unanswered.length}</span>
              </div>
            </Card>
          </section>

          <section className="space-y-4">
            <h2 className="text-xl font-bold text-ink-800">Feed</h2>

            {data.answered.map((q) => (
              <article key={q.id} className="rounded-2xl p-4 card-shadow" style={{ backgroundColor: q.background_color }}>
                <div className="mb-2 flex items-center justify-between gap-3">
                  <p className="text-sm font-semibold text-ink-800">{q.anonymous_name}</p>
                  <span className="rounded-full bg-emerald-100 px-2 py-1 text-xs font-semibold text-emerald-700">Answered</span>
                </div>
                <p className="text-sm text-ink-700">{q.content}</p>
                <div className="mt-3 rounded-xl bg-white/80 p-3">
                  <p className="text-xs font-semibold uppercase tracking-wider text-ink-500">Reply</p>
                  <p className="mt-1 text-sm text-ink-800">{q.answer}</p>
                </div>
              </article>
            ))}

            {data.unanswered.map((q) => (
              <article key={q.id} className="rounded-2xl border border-amber-200 bg-amber-50 p-4 card-shadow">
                <div className="mb-2 flex items-center justify-between gap-3">
                  <p className="text-sm font-semibold text-ink-800">{q.anonymous_name}</p>
                  <span className="rounded-full bg-amber-100 px-2 py-1 text-xs font-semibold text-amber-700">Pending</span>
                </div>
                <p className="text-sm text-ink-700">{q.content}</p>
              </article>
            ))}

            {totalCount === 0 && (
              <Card className="!p-6 bg-white text-center text-sm text-ink-500">
                No questions yet. Be the first to post one.
              </Card>
            )}
          </section>
        </div>
      )}
    </div>
  )
}
