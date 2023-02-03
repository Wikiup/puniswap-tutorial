export function shortenString(address: string, chars = 8): string {
  if (!address || address.length < 1) {
    throw Error(`Invalid address.`);
  }
  return `${address.substring(0, chars)}...${address.substring(address.length - chars)}`;
}
