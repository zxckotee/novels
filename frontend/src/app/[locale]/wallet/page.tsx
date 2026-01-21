import { unstable_setRequestLocale } from 'next-intl/server';
import { redirect } from 'next/navigation';

interface Props {
  params: { locale: string };
}

export default function WalletPage({ params: { locale } }: Props) {
  unstable_setRequestLocale(locale);
  redirect(`/${locale}/profile#wallet`);
}

