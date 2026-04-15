type DataTableControlsProps = {
  page: number;
  pageSize: number;
  total: number;
  searchValue: string;
  searchPlaceholder?: string;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  onSearchChange: (value: string) => void;
};

const pageSizeOptions = [10, 20, 50];

export function DataTableControls(props: DataTableControlsProps) {
  const totalPages = Math.max(1, Math.ceil(props.total / props.pageSize));

  return (
    <>
      <div className="table-toolbar">
        <input
          className="table-search"
          onChange={(event) => props.onSearchChange(event.target.value)}
          placeholder={props.searchPlaceholder || "输入关键词搜索"}
          value={props.searchValue}
        />
        <select
          className="table-page-size"
          onChange={(event) => props.onPageSizeChange(Number(event.target.value))}
          value={props.pageSize}
        >
          {pageSizeOptions.map((size) => (
            <option key={size} value={size}>
              每页 {size} 条
            </option>
          ))}
        </select>
      </div>
      <div className="table-pagination">
        <span className="muted">
          第 {props.page} / {totalPages} 页，共 {props.total} 条
        </span>
        <div className="action-row">
          <button
            className="table-button"
            disabled={props.page <= 1}
            onClick={() => props.onPageChange(props.page - 1)}
            type="button"
          >
            上一页
          </button>
          <button
            className="table-button"
            disabled={props.page >= totalPages}
            onClick={() => props.onPageChange(props.page + 1)}
            type="button"
          >
            下一页
          </button>
        </div>
      </div>
    </>
  );
}
