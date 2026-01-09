// API Response Types

// Novel types
export interface Novel {
  id: string;
  slug: string;
  coverUrl?: string;
  translationStatus: 'ongoing' | 'completed' | 'paused' | 'dropped';
  originalChaptersCount: number;
  releaseYear: number;
  rating: number;
  ratingsCount: number;
  viewsCount: number;
  bookmarksCount: number;
  chaptersCount: number;
  createdAt: string;
  updatedAt: string;
  // Localized fields
  title: string;
  description?: string;
  altTitles?: string[];
  // Relations
  genres?: Genre[];
  tags?: Tag[];
  author?: Author;
  latestChapter?: ChapterPreview;
}

export interface NovelListItem {
  id: string;
  slug: string;
  coverUrl?: string;
  title: string;
  rating: number;
  translationStatus: 'ongoing' | 'completed' | 'paused' | 'dropped';
  latestChapter?: number;
  updatedAt: string;
  isNew?: boolean;
}

export interface ChapterPreview {
  id: string;
  number: number;
  title: string;
  publishedAt: string;
}

// Chapter types
export interface Chapter {
  id: string;
  novelId: string;
  number: number;
  slug?: string;
  title: string;
  wordCount: number;
  publishedAt: string;
  createdAt: string;
  updatedAt: string;
}

export interface ChapterContent {
  id: string;
  chapterId: string;
  lang: string;
  content: string;
  updatedAt: string;
}

export interface ChapterWithContent extends Chapter {
  content: string;
  novel: {
    id: string;
    slug: string;
    title: string;
  };
  prevChapter?: { id: string; number: number; title: string };
  nextChapter?: { id: string; number: number; title: string };
}

// Genre and Tag types
export interface Genre {
  id: string;
  slug: string;
  name: string;
}

export interface Tag {
  id: string;
  slug: string;
  name: string;
}

export interface Author {
  id: string;
  name: string;
  originalName?: string;
}

// User types
export interface User {
  id: string;
  email: string;
  displayName: string;
  avatarUrl?: string;
  role: 'user' | 'premium' | 'moderator' | 'admin';
  level: number;
  xp: number;
  createdAt: string;
}

export interface UserProfile extends User {
  bio?: string;
  readChaptersCount: number;
  readingTime: number;
  commentsCount: number;
  bookmarksCount: number;
}

// Reading Progress
export interface ReadingProgress {
  novelId: string;
  chapterId: string;
  chapterNumber: number;
  updatedAt: string;
}

// Comment types
export interface Comment {
  id: string;
  parentId?: string;
  userId: string;
  body: string;
  likesCount: number;
  dislikesCount: number;
  repliesCount: number;
  isDeleted: boolean;
  createdAt: string;
  updatedAt: string;
  user: {
    id: string;
    displayName: string;
    avatarUrl?: string;
    level: number;
  };
  userVote?: 1 | -1;
  replies?: Comment[];
}

// Bookmark types
export interface BookmarkList {
  id: string;
  code: 'reading' | 'planned' | 'dropped' | 'completed' | 'favorites';
  title: string;
  count: number;
}

export interface Bookmark {
  id: string;
  novelId: string;
  listCode: string;
  createdAt: string;
  novel: NovelListItem;
  readingProgress?: ReadingProgress;
}

// Filter and pagination types
export interface NovelFilters {
  search?: string;
  genres?: string[];
  tags?: string[];
  status?: 'ongoing' | 'completed' | 'paused' | 'dropped';
  sort?: 'popular' | 'rating' | 'views' | 'bookmarks' | 'updated' | 'created';
  order?: 'asc' | 'desc';
  page?: number;
  limit?: number;
  lang?: string;
}

export interface PaginationMeta {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

// News types
export interface NewsPost {
  id: string;
  slug: string;
  title: string;
  body: string;
  authorId: string;
  publishedAt: string;
  createdAt: string;
  author: {
    id: string;
    displayName: string;
    avatarUrl?: string;
  };
}

// Collection types
export interface Collection {
  id: string;
  slug: string;
  title: string;
  description?: string;
  votesCount: number;
  itemsCount: number;
  createdAt: string;
  user: {
    id: string;
    displayName: string;
    avatarUrl?: string;
  };
  items?: CollectionItem[];
}

export interface CollectionItem {
  novelId: string;
  position: number;
  addedAt: string;
  novel: NovelListItem;
}
