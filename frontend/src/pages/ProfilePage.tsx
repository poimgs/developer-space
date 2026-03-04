import { type FormEvent, useEffect, useState } from 'react';
import { api, ApiError } from '../api/client';
import TagInput from '../components/TagInput';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';

const BIO_MAX = 500;
const SKILLS_MAX = 10;

const inputClass =
  'mt-1 block w-full rounded-md border border-stone-300 px-3 py-2 text-sm shadow-sm focus:border-amber-500 focus:ring-1 focus:ring-amber-500 focus:outline-none dark:border-stone-600 dark:bg-stone-800 dark:text-stone-100';

export default function ProfilePage() {
  const { user, refresh } = useAuth();
  const { addToast } = useToast();
  const [name, setName] = useState('');
  const [telegram, setTelegram] = useState('');
  const [bio, setBio] = useState('');
  const [skills, setSkills] = useState<string[]>([]);
  const [linkedinUrl, setLinkedinUrl] = useState('');
  const [instagramHandle, setInstagramHandle] = useState('');
  const [githubUsername, setGithubUsername] = useState('');
  const [saving, setSaving] = useState(false);
  const [skillSuggestions, setSkillSuggestions] = useState<string[]>([]);

  useEffect(() => {
    if (user) {
      setName(user.name);
      setTelegram(user.telegram_handle ?? '');
      setBio(user.bio ?? '');
      setSkills(user.skills ?? []);
      setLinkedinUrl(user.linkedin_url ?? '');
      setInstagramHandle(user.instagram_handle ?? '');
      setGithubUsername(user.github_username ?? '');
    }
  }, [user]);

  useEffect(() => {
    api.getSkills().then((res) => setSkillSuggestions(res.data)).catch(() => {});
  }, []);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      await api.patch('/api/auth/profile', {
        name,
        telegram_handle: telegram || null,
        bio: bio || null,
        skills,
        linkedin_url: linkedinUrl || null,
        instagram_handle: instagramHandle || null,
        github_username: githubUsername || null,
      });
      await refresh();
      addToast('Profile updated.', 'success');
    } catch (err) {
      if (err instanceof ApiError) {
        addToast(err.message, 'error');
      }
    } finally {
      setSaving(false);
    }
  };

  const bioRemaining = BIO_MAX - bio.length;

  return (
    <div className="mx-auto max-w-lg px-4 py-8">
      <h1 className="text-2xl font-bold text-stone-900 dark:text-stone-100">Profile</h1>

      {user && (
        <p className="mt-1 text-sm text-stone-500 dark:text-stone-400">
          {user.email}{user.is_admin && ' · Admin'}
        </p>
      )}

      <form onSubmit={handleSubmit} className="mt-6 space-y-8">
        {/* Identity section */}
        <section>
          <h2 className="text-lg font-semibold text-stone-900 dark:text-stone-100">Identity</h2>
          <div className="mt-4 space-y-4">
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-stone-700 dark:text-stone-300">
                Name <span className="text-red-500">*</span>
              </label>
              <input
                id="name"
                type="text"
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                className={inputClass}
              />
            </div>
          </div>
        </section>

        {/* About section */}
        <section>
          <h2 className="text-lg font-semibold text-stone-900 dark:text-stone-100">About</h2>
          <div className="mt-4 space-y-4">
            <div>
              <label htmlFor="bio" className="block text-sm font-medium text-stone-700 dark:text-stone-300">
                Bio
              </label>
              <textarea
                id="bio"
                value={bio}
                onChange={(e) => setBio(e.target.value.slice(0, BIO_MAX))}
                rows={4}
                maxLength={BIO_MAX}
                placeholder="Tell others about yourself…"
                className={inputClass}
              />
              <p
                className={`mt-1 text-right text-xs ${
                  bioRemaining <= 50
                    ? 'text-red-600 dark:text-red-400'
                    : 'text-stone-500 dark:text-stone-400'
                }`}
              >
                {bioRemaining} characters remaining
              </p>
            </div>
            <div>
              <label className="block text-sm font-medium text-stone-700 dark:text-stone-300">
                Skills
              </label>
              <div className="mt-1">
                <TagInput
                  value={skills}
                  onChange={setSkills}
                  max={SKILLS_MAX}
                  placeholder="Add a skill…"
                  suggestions={skillSuggestions}
                />
              </div>
            </div>
          </div>
        </section>

        {/* Social links section */}
        <section>
          <h2 className="text-lg font-semibold text-stone-900 dark:text-stone-100">Social Links</h2>
          <div className="mt-4 space-y-4">
            <div>
              <label htmlFor="github" className="block text-sm font-medium text-stone-700 dark:text-stone-300">
                GitHub Username
              </label>
              <input
                id="github"
                type="text"
                value={githubUsername}
                onChange={(e) => setGithubUsername(e.target.value)}
                placeholder="username"
                className={inputClass}
              />
            </div>
            <div>
              <label htmlFor="linkedin" className="block text-sm font-medium text-stone-700 dark:text-stone-300">
                LinkedIn URL
              </label>
              <input
                id="linkedin"
                type="url"
                value={linkedinUrl}
                onChange={(e) => setLinkedinUrl(e.target.value)}
                placeholder="https://linkedin.com/in/username"
                className={inputClass}
              />
            </div>
            <div>
              <label htmlFor="telegram" className="block text-sm font-medium text-stone-700 dark:text-stone-300">
                Telegram Handle
              </label>
              <input
                id="telegram"
                type="text"
                value={telegram}
                onChange={(e) => setTelegram(e.target.value)}
                placeholder="@username"
                className={inputClass}
              />
            </div>
            <div>
              <label htmlFor="instagram" className="block text-sm font-medium text-stone-700 dark:text-stone-300">
                Instagram Handle
              </label>
              <input
                id="instagram"
                type="text"
                value={instagramHandle}
                onChange={(e) => setInstagramHandle(e.target.value)}
                placeholder="username"
                className={inputClass}
              />
            </div>
          </div>
        </section>

        <div className="flex justify-end">
          <button
            type="submit"
            disabled={saving}
            className="rounded-md bg-amber-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-amber-700 focus:ring-2 focus:ring-amber-500 focus:outline-none disabled:cursor-not-allowed disabled:bg-stone-300 dark:disabled:bg-stone-700"
          >
            {saving ? 'Saving...' : 'Save'}
          </button>
        </div>
      </form>
    </div>
  );
}
