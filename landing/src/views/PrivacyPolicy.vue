<template>
  <div class="min-h-screen bg-coal-900 font-body">
    <LandingNav />

    <main class="max-w-3xl mx-auto px-6 pt-32 pb-24">
      <!-- Header -->
      <div class="mb-12">
        <div class="inline-flex items-center gap-2 bg-coal-800 border border-ember/20 rounded-full px-4 py-1.5 mb-6">
          <span class="w-1.5 h-1.5 rounded-full bg-ember" />
          <span class="text-xs font-semibold text-ember tracking-widest uppercase">Legal</span>
        </div>
        <h1 class="font-display font-800 text-4xl md:text-5xl text-white mb-4">Privacy Policy</h1>
        <p class="text-white/40 text-sm">Last updated: March 1, 2026</p>
      </div>

      <!-- Content -->
      <div class="prose-legal">
        <p>
          chatcoal ("we", "us", or "our") is an open-source messaging platform. This Privacy Policy explains what information we collect, how we use it, and your rights regarding that information when you use our managed service at <a href="https://app.chatcoal.com">app.chatcoal.com</a>.
        </p>
        <p>
          If you self-host chatcoal, this policy does not apply — you are the data controller and are responsible for your users' data.
        </p>

        <h2>1. Information We Collect</h2>
        <h3>Account Information</h3>
        <p>When you register, we collect your username, email address, and a hashed password. We do not store your password in plaintext.</p>

        <h3>Messages and Content</h3>
        <p>Messages you send, files you upload, and other content you create are stored on our servers to provide the service. Message content is not end-to-end encrypted on the managed service — server administrators can access content in channels they manage.</p>

        <h3>Usage Data</h3>
        <p>We collect basic usage metrics (e.g. number of messages sent, feature usage) to understand how the product is used and to improve it. This data is aggregated and not linked to individual users.</p>

        <h3>Technical Data</h3>
        <p>We log IP addresses and user agent strings for security and abuse prevention purposes. These logs are retained for 30 days.</p>

        <h2>2. How We Use Your Information</h2>
        <ul>
          <li>To provide, maintain, and improve the chatcoal service</li>
          <li>To authenticate your account and keep it secure</li>
          <li>To send transactional emails (e.g. password resets) — we do not send marketing emails</li>
          <li>To detect and prevent abuse, spam, and violations of our Terms of Service</li>
          <li>To comply with legal obligations</li>
        </ul>

        <h2>3. Data Sharing</h2>
        <p>We do not sell your personal data. We do not share your data with third parties for advertising purposes. We may share data:</p>
        <ul>
          <li>With infrastructure providers (e.g. hosting) strictly to operate the service</li>
          <li>When required by law or valid legal process</li>
          <li>To protect the rights and safety of chatcoal and its users</li>
        </ul>

        <h2>4. Data Retention</h2>
        <p>Your account data is retained as long as your account exists. If you delete your account, your profile and messages are deleted within 30 days. Some data (e.g. security logs) may be retained for up to 90 days after deletion.</p>

        <h2>5. Open Source</h2>
        <p>chatcoal is fully open source. You can review exactly how the software handles your data by reading the source code on GitHub. There is no hidden telemetry or tracking beyond what is described in this policy.</p>

        <h2>6. Your Rights</h2>
        <p>Depending on your location, you may have rights including:</p>
        <ul>
          <li>Access to the personal data we hold about you</li>
          <li>Correction of inaccurate data</li>
          <li>Deletion of your account and associated data</li>
          <li>Portability of your data</li>
        </ul>
        <p>To exercise these rights, delete your account from the app settings or contact us at the address below.</p>

        <h2>7. Cookies</h2>
        <p>We use a single session cookie to keep you logged in. We do not use tracking cookies, analytics cookies, or third-party advertising cookies.</p>

        <h2>8. Children</h2>
        <p>chatcoal is not directed at children under 13. If you believe a child has created an account, please contact us and we will delete it promptly.</p>

        <h2>9. Changes to This Policy</h2>
        <p>We may update this policy. We will notify users of significant changes by posting a notice in the app. Continued use after changes constitutes acceptance.</p>

        <h2>10. Contact</h2>
        <p>Questions about this policy? Open an issue on our GitHub repository or reach out through the app.</p>
      </div>
    </main>

    <LandingFooter />
  </div>
</template>

<script setup>
import LandingNav from '../components/LandingNav.vue'
import LandingFooter from '../components/LandingFooter.vue'
</script>

<style scoped>
.prose-legal {
  color: rgba(255, 255, 255, 0.55);
  font-size: 0.9375rem;
  line-height: 1.8;
}

.prose-legal p {
  margin-bottom: 1.25rem;
}

.prose-legal h2 {
  font-family: var(--font-display, 'Syne', sans-serif);
  font-weight: 700;
  font-size: 1.25rem;
  color: rgba(255, 255, 255, 0.9);
  margin-top: 2.5rem;
  margin-bottom: 0.75rem;
}

.prose-legal h3 {
  font-weight: 600;
  font-size: 1rem;
  color: rgba(255, 255, 255, 0.7);
  margin-top: 1.5rem;
  margin-bottom: 0.5rem;
}

.prose-legal ul {
  list-style: none;
  padding: 0;
  margin-bottom: 1.25rem;
}

.prose-legal ul li {
  padding-left: 1.25rem;
  position: relative;
  margin-bottom: 0.4rem;
}

.prose-legal ul li::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0.65em;
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: #E8521A;
  opacity: 0.7;
}

.prose-legal a {
  color: #E8521A;
  text-decoration: none;
}

.prose-legal a:hover {
  text-decoration: underline;
}
</style>
