import { getRequestConfig } from 'next-intl/server';
import { notFound } from 'next/navigation';
import { locales, type Locale } from './i18n/config';

export default getRequestConfig(async ({ locale }) => {
  // Validate that the incoming `locale` parameter is valid
  if (!locales.includes(locale as Locale)) {
    notFound();
  }

  // Load base messages - fallback to en if locale doesn't exist
  let baseMessages: any;
  try {
    // Use dynamic import with explicit .json extension
    const localeModule = await import(`../messages/${locale}.json`);
    baseMessages = localeModule.default || localeModule;
  } catch (error) {
    // Fallback to English if locale file doesn't exist
    try {
      const enModule = await import(`../messages/en.json`);
      baseMessages = enModule.default || enModule;
    } catch {
      // If even English fails, use empty object
      baseMessages = {};
    }
  }
  
  // Load additional namespace messages if they exist
  const additionalMessages: Record<string, any> = {};
  
  // Try to load from locale-specific directory, fallback to en
  const loadNamespace = async (namespace: string) => {
    try {
      // NOTE: Namespace messages live under src/messages/, not messages/
      const module = await import(`./messages/${locale}/${namespace}.json`);
      return module.default || module;
    } catch {
      try {
        const enModule = await import(`./messages/en/${namespace}.json`);
        return enModule.default || enModule;
      } catch {
        return {};
      }
    }
  };
  
  additionalMessages.community = await loadNamespace('community');
  const economyMessages = await loadNamespace('economy');
  additionalMessages.moderation = await loadNamespace('moderation');

  // Merge economy messages at root level (voting, proposals, wallet are top-level keys in economy.json)
  // This allows useTranslations('voting') to work directly
  const mergedMessages: Record<string, any> = {
    ...baseMessages,
    ...additionalMessages,
    // Spread economy messages at root level so voting, proposals, wallet are accessible directly
    ...(economyMessages || {}),
    // Keep economy namespace too for backwards compatibility if needed
    economy: economyMessages,
  };

  return {
    messages: mergedMessages,
  };
});
