Requires [Go 1.22+](https://golang.org/dl/)

### EverLast

Since the Evmos repo licensed, we’ve forked it to [Evermint](https://github.com/EscanBE/evermint) to keep building and enhancing the codebase while keeping it opensource.
And EverLast is a derivative products of Evermint.

EverLast was born with three main purposes:
- Revitalizing the EVM-Enabled Blockchain Experience.
- Building a fair-launch, community-driven home for EVM-Enabled Blockchain users.
- Battle test the new codebase.

Built on Evmos v12.1.6, Evermint introduces significant enhancements and new innovations to empower users and developers alike, and EverLast introduces those to the community.
Evermint support [stateful precompiled contracts](https://github.com/EscanBE/evermint/pull/175) in a real way to allow expanding to all Cosmos-SDK modules with minimum effort.

### Snapshot from Evmos

To support the Evmos community, we are providing a migration path for Evmos users to EverLast by capturing the current balance of users on Evmos into EverLast. _(A deduction factor is applied during migration to ensure a fair distribution and mitigate the disproportionate holdings previously accumulated by foundation of the another chain.)_

- Snapshot 1 at block 21302452 at June 5th, 2024 ([snapshot code](https://github.com/EscanBE/fork-evmos-for-snapshot/tree/snapshot/v18.1.0))
- Snapshot 2 at block 28318578 at April 14th, 2025 ([snapshot code](https://github.com/EscanBE/fork-evmos-for-snapshot/tree/snapshot/v20.0.0))

Total 3,249,247 EVL for users from both snapshots, after deduction factor applied.