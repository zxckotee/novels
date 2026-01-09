# UI Компоненты и Дизайн-система

## Обзор дизайн-системы

### Цветовая схема (Dark Theme)
```css
:root {
  /* Primary Colors */
  --color-primary: #6366f1; /* Indigo для акцентов и кнопок */
  --color-primary-hover: #5856eb;
  --color-primary-light: rgba(99, 102, 241, 0.1);
  
  /* Background */
  --bg-primary: #121212; /* Основной темный фон */
  --bg-secondary: #1e1e1e; /* Карточки, модалы */
  --bg-tertiary: #2a2a2a; /* Хедер, сайдбар */
  --bg-hover: #333333; /* Ховер состояния */
  
  /* Text Colors */
  --text-primary: #ffffff; /* Основной текст */
  --text-secondary: #a1a1a1; /* Второстепенный текст */
  --text-tertiary: #666666; /* Подписи, метаданные */
  
  /* Border & Dividers */
  --border-color: #333333;
  --border-light: #2a2a2a;
  
  /* Status Colors */
  --color-success: #10b981; /* Завершено */
  --color-warning: #f59e0b; /* В процессе */
  --color-error: #ef4444; /* Брошено/ошибки */
  --color-info: #3b82f6; /* Информация */
  
  /* Градиенты */
  --gradient-primary: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  --gradient-card: linear-gradient(135deg, #1e1e1e 0%, #2a2a2a 100%);
  
  /* Shadows */
  --shadow-small: 0 1px 3px rgba(0, 0, 0, 0.4);
  --shadow-medium: 0 4px 6px rgba(0, 0, 0, 0.5);
  --shadow-large: 0 10px 25px rgba(0, 0, 0, 0.6);
  
  /* Border Radius */
  --radius-small: 4px;
  --radius-medium: 8px;
  --radius-large: 12px;
  --radius-xl: 16px;
  
  /* Spacing Scale */
  --space-1: 0.25rem; /* 4px */
  --space-2: 0.5rem;  /* 8px */
  --space-3: 0.75rem; /* 12px */
  --space-4: 1rem;    /* 16px */
  --space-5: 1.25rem; /* 20px */
  --space-6: 1.5rem;  /* 24px */
  --space-8: 2rem;    /* 32px */
  --space-10: 2.5rem; /* 40px */
  --space-12: 3rem;   /* 48px */
  --space-16: 4rem;   /* 64px */
}
```

### Типографика
```css
/* Font Families */
--font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
--font-mono: 'JetBrains Mono', 'Fira Code', monospace;

/* Font Sizes */
--text-xs: 0.75rem;    /* 12px */
--text-sm: 0.875rem;   /* 14px */
--text-base: 1rem;     /* 16px */
--text-lg: 1.125rem;   /* 18px */
--text-xl: 1.25rem;    /* 20px */
--text-2xl: 1.5rem;    /* 24px */
--text-3xl: 1.875rem;  /* 30px */
--text-4xl: 2.25rem;   /* 36px */

/* Line Heights */
--leading-tight: 1.25;
--leading-normal: 1.5;
--leading-relaxed: 1.75;

/* Font Weights */
--font-light: 300;
--font-normal: 400;
--font-medium: 500;
--font-semibold: 600;
--font-bold: 700;
```

## Базовые UI компоненты

### Button
```typescript
interface ButtonProps {
  variant: 'primary' | 'secondary' | 'ghost' | 'danger';
  size: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  loading?: boolean;
  icon?: ReactNode;
  children: ReactNode;
  onClick?: () => void;
}

// Стили по вариантам
const buttonVariants = {
  primary: 'bg-primary text-white hover:bg-primary-hover',
  secondary: 'bg-bg-secondary text-text-primary border border-border-color hover:bg-bg-hover',
  ghost: 'text-text-primary hover:bg-bg-hover',
  danger: 'bg-error text-white hover:bg-red-600'
};

const buttonSizes = {
  sm: 'px-3 py-1.5 text-sm',
  md: 'px-4 py-2 text-base', 
  lg: 'px-6 py-3 text-lg'
};
```

### Card
```typescript
interface CardProps {
  variant?: 'default' | 'hover' | 'bordered';
  padding?: 'sm' | 'md' | 'lg';
  children: ReactNode;
  className?: string;
}

// Базовые стили карточки
const cardStyles = `
  background: var(--bg-secondary);
  border-radius: var(--radius-large);
  box-shadow: var(--shadow-medium);
  border: 1px solid var(--border-color);
`;

// Варианты карточек
const cardVariants = {
  default: '',
  hover: 'transition-transform hover:scale-105 cursor-pointer',
  bordered: 'border-2 border-primary'
};
```

### Input & Form Components
```typescript
interface InputProps {
  type?: 'text' | 'email' | 'password' | 'search';
  placeholder?: string;
  value?: string;
  onChange?: (value: string) => void;
  disabled?: boolean;
  error?: string;
  icon?: ReactNode;
  size?: 'sm' | 'md' | 'lg';
}

interface SelectProps {
  options: { value: string; label: string }[];
  value?: string;
  onChange?: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;
  multiple?: boolean;
}

interface TextareaProps {
  value?: string;
  onChange?: (value: string) => void;
  placeholder?: string;
  rows?: number;
  disabled?: boolean;
  error?: string;
}
```

### Modal & Dialog
```typescript
interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  size?: 'sm' | 'md' | 'lg' | 'xl';
  children: ReactNode;
}

interface ConfirmDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  variant?: 'default' | 'danger';
}
```

### Dropdown & Tooltip
```typescript
interface DropdownProps {
  trigger: ReactNode;
  children: ReactNode;
  position?: 'bottom-left' | 'bottom-right' | 'top-left' | 'top-right';
  disabled?: boolean;
}

interface TooltipProps {
  content: string | ReactNode;
  position?: 'top' | 'bottom' | 'left' | 'right';
  children: ReactNode;
}
```

## Специализированные компоненты

### NovelCard (Карточка новеллы)
```typescript
interface NovelCardProps {
  novel: Novel;
  variant?: 'grid' | 'list' | 'slider';
  showProgress?: boolean;
  showBookmarkButton?: boolean;
  onBookmark?: (novel: Novel) => void;
  onClick?: (novel: Novel) => void;
}

// Структура компонента:
// - Cover image (обложка)
// - Title (название)
// - Author (автор)  
// - Rating (звездочки + число)
// - Status badge (продолжается/завершен)
// - Tags (первые 3 тега)
// - Bookmark button (если авторизован)
// - Progress bar (если есть прогресс)
```

### ChapterCard (Карточка главы)
```typescript
interface ChapterCardProps {
  chapter: Chapter;
  isRead?: boolean;
  isNew?: boolean;
  showDate?: boolean;
  onClick?: (chapter: Chapter) => void;
}

// Структура:
// - Chapter number & title
// - Publication date
// - Read status indicator
// - New chapter badge
// - Comments count
// - Reading time estimate
```

### CommentThread (Древовидные комментарии)
```typescript
interface CommentThreadProps {
  comments: Comment[];
  maxDepth?: number;
  onReply?: (parentId: string, body: string) => void;
  onVote?: (commentId: string, value: 1 | -1) => void;
  onReport?: (commentId: string, reason: string) => void;
  onEdit?: (commentId: string, body: string) => void;
  onDelete?: (commentId: string) => void;
}

interface CommentItemProps {
  comment: Comment;
  depth: number;
  children?: ReactNode;
}

// Структура Comment:
// - User info (avatar, name, level badge)
// - Comment body
// - Actions (reply, vote, report, edit, delete)
// - Nested replies (if depth < maxDepth)
// - Vote score
// - Timestamp
```

### Reader (Ридер для чтения глав)
```typescript
interface ReaderProps {
  chapter: ChapterContent;
  settings: ReaderSettings;
  onSettingsChange: (settings: ReaderSettings) => void;
  onProgress: (position: number) => void;
  onNavigate: (direction: 'prev' | 'next' | 'chapters') => void;
}

interface ReaderSettings {
  fontSize: number; // 14-24px
  lineHeight: number; // 1.4-2.0
  fontFamily: 'serif' | 'sans-serif' | 'mono';
  backgroundColor: string;
  textColor: string;
  maxWidth: number; // 600-1200px
  padding: number;
}

// Функции ридера:
// - Настройки типографики
// - Прогресс-бар чтения
// - Навигация между главами
// - Абзацные комментарии (позже)
// - Закладки в тексте
// - Поделиться цитатой
```

### BookmarkManager (Управление закладками)
```typescript
interface BookmarkManagerProps {
  bookmarks: BookmarkItem[];
  lists: BookmarkList[];
  currentFilter: string;
  onFilterChange: (filter: string) => void;
  onMove: (novelId: string, newListId: string) => void;
  onRemove: (novelId: string) => void;
  onSort: (sortBy: 'updated_at' | 'created_at' | 'title') => void;
}

interface BookmarkItemProps {
  item: BookmarkItem;
  onMove: (listId: string) => void;
  onRemove: () => void;
  onClick: () => void;
}

// Структура:
// - Filter tabs (Читаю, Планы, Брошено и т.д.)
// - Sort dropdown
// - Grid/List view toggle  
// - Bookmark items с прогрессом
// - "Continue reading" кнопка
// - Context menu для перемещения
```

### VotingInterface (Голосование)
```typescript
interface VotingInterfaceProps {
  poll: VotingPoll;
  userBalances: TicketBalance[];
  onVote: (proposalId: string, ticketType: string, amount: number) => void;
}

interface ProposalCardProps {
  proposal: NovelProposal;
  currentScore: number;
  position: number;
  onVote: (ticketType: string, amount: number) => void;
  userBalances: TicketBalance[];
}

// Структура:
// - Timer countdown до окончания голосования
// - Список предложений с рейтингом
// - Voting actions (Daily Vote / Translation Ticket)
// - User balance display
// - History of user votes
```

### TicketWallet (Кошелек билетов)
```typescript
interface TicketWalletProps {
  balances: TicketBalance[];
  transactions: TicketTransaction[];
  onViewHistory: () => void;
}

interface TicketBalanceCardProps {
  balance: TicketBalance;
  description: string;
  nextUpdate?: string;
}

// Отображение в хедере:
// - Daily Vote count с иконкой
// - Novel Request count
// - Translation Ticket count  
// - Dropdown с деталями и историей
```

## Layout компоненты

### Header (Шапка сайта)
```typescript
interface HeaderProps {
  isAuthenticated: boolean;
  user?: User;
  currentLocale: string;
  onLocaleChange: (locale: string) => void;
  onLogout: () => void;
}

// Структура (как на дизайне):
// - Logo (слева)
// - Search bar (центр)
// - Navigation links (Каталог, Топы, Форум, Новости)
// - User controls (справа):
//   - Ticket balances
//   - Notifications bell
//   - Bookmarks icon  
//   - User avatar & dropdown
//   - Language switcher
```

### Sidebar (Боковая панель)
```typescript
interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
  user?: User;
}

// Мобильная навигация:
// - User info (если авторизован)
// - Main navigation links
// - Quick access to bookmarks
// - Settings & logout
```

### Footer
```typescript
interface FooterProps {
  currentLocale: string;
}

// Структура:
// - Links (О проекте, Помощь, API)
// - Social media
// - Language switcher (дублирует хедер)
// - Copyright & version
```

### PageLayout
```typescript
interface PageLayoutProps {
  title?: string;
  description?: string;
  noIndex?: boolean;
  children: ReactNode;
}

// SEO-обертка для страниц:
// - Head meta tags
// - OpenGraph data 
// - JSON-LD structured data
// - Canonical URLs
```

## Страницы и их компоненты

### HomePage
```typescript
// Компоненты главной страницы:
// - HeroSlider (рекомендации)
// - QuickActions (Голосовать, Предложить)
// - NewsWidget (последние новости)
// - TopSections (Популярное, Новинки, Тренды)
// - RecentUpdates (последние обновления)
// - PopularCollections (коллекции пользователей)

interface HeroSliderProps {
  recommendations: Novel[];
  autoPlay?: boolean;
  interval?: number;
}

interface TopSectionProps {
  title: string;
  novels: Novel[];
  viewAllLink: string;
  variant?: 'grid' | 'scroll';
}

interface RecentUpdatesProps {
  updates: {
    novel: Novel;
    chapter: Chapter;
    timeAgo: string;
  }[];
  limit?: number;
}
```

### CatalogPage
```typescript
// Компоненты каталога:
// - SearchBar с автодополнением
// - FilterPanel (жанры, теги, статус, год)
// - SortDropdown
// - ViewToggle (grid/list)
// - NovelGrid/NovelList
// - Pagination

interface FilterPanelProps {
  genres: Genre[];
  tags: Tag[];
  selectedFilters: FilterState;
  onFilterChange: (filters: FilterState) => void;
  onReset: () => void;
}

interface FilterState {
  genres: string[];
  tags: string[];
  status: string[];
  yearRange: [number, number];
  rating: number;
}
```

### NovelPage
```typescript
// Компоненты страницы новеллы:
// - NovelHeader (обложка, info, actions)
// - TabNavigation (Описание, Главы, Комментарии)
// - DescriptionTab
// - ChaptersTab (список глав с пагинацией)
// - CommentsTab
// - BookmarkDropdown
// - RatingWidget

interface NovelHeaderProps {
  novel: NovelDetail;
  onBookmark: (listCode: string) => void;
  onRate: (rating: number) => void;
  canEdit: boolean;
}

interface ChaptersTabProps {
  chapters: Chapter[];
  userProgress?: ReadingProgress;
  onChapterClick: (chapter: Chapter) => void;
}
```

### ReaderPage
```typescript
// Компоненты ридера:
// - ReaderHeader (навигация, настройки)
// - ReaderContent (текст главы)
// - ReaderFooter (прогресс, навигация)
// - SettingsPanel (типографика)
// - CommentsPanel (обычные + абзацные)

interface ReaderHeaderProps {
  novel: Novel;
  chapter: Chapter;
  onBack: () => void;
  onShowChapters: () => void;
  onShowSettings: () => void;
}

interface ReaderContentProps {
  content: string;
  settings: ReaderSettings;
  onProgress: (position: number) => void;
  onSelectionComment?: (selection: TextSelection) => void;
}
```

### ProfilePage
```typescript
// Компоненты профиля:
// - ProfileHeader (avatar, stats, level)
// - TabNavigation (Активность, Закладки, Комментарии, Коллекции)
// - ActivityTab (история чтения, достижения)
// - BookmarksTab 
// - CommentsTab
// - CollectionsTab
// - SettingsTab (только для своего профиля)

interface ProfileHeaderProps {
  profile: UserProfile;
  isOwner: boolean;
  onEdit?: () => void;
}

interface LevelProgressProps {
  level: number;
  xp: number;
  xpToNext: number;
}
```

## Адаптивность и мобильная версия

### Breakpoints
```css
/* Mobile First подход */
--mobile: 320px;
--mobile-lg: 480px;
--tablet: 768px;
--desktop: 1024px;
--desktop-lg: 1280px;
--desktop-xl: 1536px;
```

### Мобильные особенности
```typescript
// Mobile-specific компоненты:
// - SwipeableCard (свайп для действий)
// - BottomSheetModal (модалы снизу)
// - TabBarNavigation (нижняя навигация)
// - PullToRefresh
// - InfiniteScroll

interface SwipeableCardProps {
  children: ReactNode;
  onSwipeLeft?: () => void;
  onSwipeRight?: () => void;
  leftAction?: { icon: ReactNode; color: string };
  rightAction?: { icon: ReactNode; color: string };
}
```

## Анимации и переходы

### Стандартные анимации
```css
/* Transitions */
--transition-fast: 0.15s ease-out;
--transition-normal: 0.3s ease-out;
--transition-slow: 0.5s ease-out;

/* Animations */
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes slideIn {
  from { transform: translateX(-100%); }
  to { transform: translateX(0); }
}

@keyframes scaleIn {
  from { transform: scale(0.9); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}

/* Loading animations */
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
```

## Accessibility

### ARIA и семантика
```typescript
// Обязательные accessibility атрибуты:
// - aria-label для иконочных кнопок
// - aria-expanded для dropdown/collapse
// - aria-describedby для error messages
// - role для кастомных компонентов
// - tabIndex для keyboard navigation
// - focus management для модалов

// Семантические HTML элементы:
// - <main> для основного контента
// - <nav> для навигации
// - <article> для новелл/глав
// - <aside> для сайдбара
// - <section> для логических блоков
// - <header>/<footer> для шапки/подвала
```

### Клавиатурная навигация
```typescript
// Горячие клавиши:
// - Ctrl+K / Cmd+K: открыть поиск
// - Escape: закрыть модалы/dropdown
// - Arrow keys: навигация в меню
// - Enter/Space: активация кнопок
// - Tab/Shift+Tab: переход между элементами

// В ридере:
// - Arrow Left/Right: предыдущая/следующая глава  
// - Arrow Up/Down: прокрутка
// - S: открыть настройки
// - C: открыть комментарии
// - B: добавить в закладки
```

## Производительность

### Оптимизации
```typescript
// Lazy loading компонентов:
const CommentThread = lazy(() => import('./CommentThread'));
const ReaderSettings = lazy(() => import('./ReaderSettings'));

// Мемоизация тяжелых вычислений:
const MemoizedNovelCard = memo(NovelCard);
const MemoizedCommentItem = memo(CommentItem);

// Виртуализация длинных списков:
import { FixedSizeList as List } from 'react-window';

// Intersection Observer для lazy loading изображений:
const LazyImage = ({ src, alt, ...props }) => {
  const [loaded, setLoaded] = useState(false);
  const ref = useIntersectionObserver(() => setLoaded(true));
  
  return loaded ? <img src={src} alt={alt} {...props} /> : <div ref={ref} />;
};
```

Эта дизайн-система обеспечивает консистентность UI/UX по всему приложению и упрощает разработку за счет переиспользуемых компонентов.