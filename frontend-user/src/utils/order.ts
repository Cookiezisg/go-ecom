export function getOrderStatusText(status: number) {
  switch (status) {
    case 0:
      return "已取消";
    case 1:
      return "待支付";
    case 2:
      return "待发货";
    case 3:
      return "待收货";
    case 4:
      return "已完成";
    case 5:
      return "已退款";
    default:
      return "未知状态";
  }
}
