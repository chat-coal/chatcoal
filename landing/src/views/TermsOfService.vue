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
        <h1 class="font-display font-800 text-4xl md:text-5xl text-white mb-4">Terms of Service</h1>
        <p class="text-white/40 text-sm">Last updated: March 1, 2026</p>
      </div>

      <!-- Content -->
      <div class="prose-legal">
        <p>
          By accessing or using chatcoal's managed service at <a href="https://app.chatcoal.com">app.chatcoal.com</a>, you agree to these Terms of Service. If you do not agree, do not use the service.
        </p>

        <h2>1. The Service</h2>
        <p>chatcoal provides real-time messaging infrastructure including servers, channels, direct messages, and related features. The service is provided as-is and may be updated, modified, or discontinued at any time.</p>
        <p>chatcoal is open-source software. You may self-host it under the terms of the project's open-source license. These Terms apply only to the managed service we operate.</p>

        <h2>2. Accounts</h2>
        <ul>
          <li>You must be at least 13 years old to create an account</li>
          <li>You are responsible for maintaining the security of your account credentials</li>
          <li>You are responsible for all activity that occurs under your account</li>
          <li>You may not create accounts for the purpose of abuse, spam, or impersonation</li>
          <li>One person or entity may not maintain multiple accounts to circumvent restrictions</li>
        </ul>

        <h2>3. Acceptable Use</h2>
        <p>You agree not to use chatcoal to:</p>
        <ul>
          <li>Violate any applicable law or regulation</li>
          <li>Harass, threaten, or harm other users</li>
          <li>Distribute spam, malware, or phishing content</li>
          <li>Post illegal content including CSAM — violations will be reported to authorities</li>
          <li>Attempt to gain unauthorized access to other accounts or systems</li>
          <li>Scrape, crawl, or systematically extract data from the service</li>
          <li>Use the service to facilitate illegal activity</li>
        </ul>

        <h2>4. Content</h2>
        <p>You retain ownership of content you create. By posting content, you grant chatcoal a limited license to store and transmit it solely to provide the service. We do not claim ownership of your content and do not use it for advertising.</p>
        <p>You are solely responsible for the content you post. We reserve the right to remove content that violates these Terms.</p>

        <h2>5. Server Owners and Administrators</h2>
        <p>If you create a server, you are responsible for moderating it in accordance with these Terms. Hosting illegal content or facilitating abuse through your server may result in termination of both the server and your account.</p>

        <h2>6. Termination</h2>
        <p>We may suspend or terminate your account at any time if you violate these Terms. You may delete your account at any time from the app settings. Upon termination, your data will be deleted in accordance with our Privacy Policy.</p>

        <h2>7. Service Availability</h2>
        <p>We strive for high availability but make no guarantees. The service may be unavailable due to maintenance, outages, or circumstances beyond our control. We are not liable for any loss resulting from service interruptions.</p>

        <h2>8. Disclaimer of Warranties</h2>
        <p>The service is provided "as is" without warranties of any kind, express or implied. We do not warrant that the service will be error-free, secure, or available at any particular time.</p>

        <h2>9. Limitation of Liability</h2>
        <p>To the maximum extent permitted by law, chatcoal shall not be liable for any indirect, incidental, special, consequential, or punitive damages arising from your use of or inability to use the service.</p>

        <h2>10. Changes to These Terms</h2>
        <p>We may update these Terms. We will notify users of material changes by posting a notice in the app. Continued use after changes constitutes acceptance of the updated Terms.</p>

        <h2>11. Contact</h2>
        <p>Questions about these Terms? Open an issue on our GitHub repository or reach out through the app.</p>
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
