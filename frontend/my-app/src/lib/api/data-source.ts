export type DataSource = 'mock' | 'api';

export const dataSource: DataSource =
  process.env.NEXT_PUBLIC_DATA_SOURCE === 'api' ? 'api' : 'mock';