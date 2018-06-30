package netaddr


//A function that accepts an iterable sequence of IP addresses and subnets
//merging them into the smallest possible list of CIDRs. It merges adjacent
//subnets where possible, those contained within others and also removes
//any duplicates.
//:param ip_addrs: an iterable sequence of IP addresses and subnets.
//:return: a summarized list of `IPNetwork` objects.
func CidrMerge(networks ...IPNetwork) (summarised *[]IPNetwork) {
	// The algorithm is quite simple: For each CIDR we create an IP range.
	// Sort them and merge when possible.  Afterwards split them again
	// optimally.
	summarised := new([]IPNetwork)

	for _, ip := range networks {
		summarised = append(summarised, )
	}



	return networks
}

ranges = []

for ip in ip_addrs:
	cidr = IPNetwork(ip)
	# Since non-overlapping ranges are the common case, remember the original
	ranges.append( (cidr.version, cidr.last, cidr.first, cidr) )

ranges.sort()
i = len(ranges) - 1
while i > 0:
	if ranges[i][0] == ranges[i - 1][0] and ranges[i][2] - 1 <= ranges[i - 1][1]:
		ranges[i - 1] = (ranges[i][0], ranges[i][1], min(ranges[i - 1][2], ranges[i][2]))
		del ranges[i]
	i -= 1
merged = []
for range_tuple in ranges:
# If this range wasn't merged we can simply use the old cidr.
	if len(range_tuple) == 4:
		merged.append(range_tuple[3])
	else:
		version = range_tuple[0]
		range_start = IPAddress(range_tuple[2], version=version)
		range_stop = IPAddress(range_tuple[1], version=version)
		merged.extend(iprange_to_cidrs(range_start, range_stop))
return merged
