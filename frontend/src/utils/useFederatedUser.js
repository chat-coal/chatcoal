/**
 * useFederatedUser — returns display helpers for a user object.
 *
 * Usage:
 *   const { displayName, instanceBadge } = useFederatedUser(user)
 *
 * `instanceBadge` is e.g. "@other.com" when the user is federated, or null.
 */
export function useFederatedUser(user) {
  const displayName = user?.display_name || 'Unknown'
  const instanceBadge = user?.home_instance ? `@${user.home_instance}` : null
  return { displayName, instanceBadge }
}
