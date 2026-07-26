import { useAuth } from '../hooks/useAuth'
import Button from '../components/ui/Button'

export default function LoginPage() {
  const { login } = useAuth()

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden px-4 py-10">
      <div className="mesh-bg card-shadow soft-outline absolute inset-0 opacity-80" />
      <div className="relative z-10 w-full max-w-5xl overflow-hidden rounded-3xl bg-white/90 shadow-2xl backdrop-blur">
        <div className="grid md:grid-cols-2">
          <section className="bg-gradient-to-br from-brand-500 to-brand-700 p-8 text-white md:p-10">
            <p className="text-sm font-semibold tracking-[0.2em] text-brand-100">POSTCUBE</p>
            <h1 className="mt-4 text-4xl leading-tight font-bold">Collect anonymous questions in your own box.</h1>
            <p className="mt-5 text-base/7 text-brand-50">
              Share one link. Receive anonymous questions. Reply from your private inbox and publish both answered and pending messages on your profile page.
            </p>
            <div className="mt-8 rounded-2xl bg-white/15 p-4 text-sm text-brand-50">
              Friendly flow: public ask page + private inbox + background color styling per question card.
            </div>
          </section>

          <section className="p-8 md:p-10">
            <h2 className="text-3xl font-bold text-ink-800">Sign in</h2>
            <p className="mt-2 text-ink-500">Use your BasaltPass account to manage your question box.</p>

            <Button variant="dark" fullWidth className="mt-8" onClick={login}>
              Continue with BasaltPass
            </Button>

            <div className="mt-8 rounded-xl border border-ink-100 bg-ink-50 p-4 text-sm text-ink-600">
              New users are auto-provisioned with a unique question box URL after OAuth callback.
            </div>
          </section>
        </div>
      </div>
    </div>
  )
}
