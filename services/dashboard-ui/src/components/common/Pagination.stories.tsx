import { Pagination } from './Pagination'

export const Default = () => <Pagination />

export const WithData = () => <Pagination hasNext={true} offset={10} />
