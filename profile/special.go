package profile

import (
	"time"

	"github.com/safing/portbase/log"
)

const (
	// UnidentifiedProfileID is the profile ID used for unidentified processes.
	UnidentifiedProfileID = "_unidentified"
	// UnidentifiedProfileName is the name used for unidentified processes.
	UnidentifiedProfileName = "Unidentified Processes"
	// UnidentifiedProfileDescription is the description used for unidentified processes.
	UnidentifiedProfileDescription = `This is not a real application, but a collection of connections that could not be attributed to a process. This could be because the Portmaster failed to identify the process, or simply because there is no process waiting for an incoming connection.

Seeing a lot of incoming connections here is normal, as this resembles the network chatter of other devices.
`

	// SystemProfileID is the profile ID used for the system/kernel.
	SystemProfileID = "_system"
	// SystemProfileName is the name used for the system/kernel.
	SystemProfileName = "Operating System"
	// SystemProfileDescription is the description used for the system/kernel.
	SystemProfileDescription = "This is the operation system itself."

	// SystemResolverProfileID is the profile ID used for the system's DNS resolver.
	SystemResolverProfileID = "_system-resolver"
	// SystemResolverProfileName is the name used for the system's DNS resolver.
	SystemResolverProfileName = "System DNS Client"
	// SystemResolverProfileDescription is the description used for the system's DNS resolver.
	SystemResolverProfileDescription = `The System DNS Client is a system service that requires special handling. For regular network connections, the configured settings will apply as usual, but DNS requests coming from the System DNS Client are handled in a special way, as they could actually be coming from any other application on the system.

In order to respect the app settings of the actual application, DNS requests from the System DNS Client are only subject to the following settings:

- Outgoing Rules (without global rules)
- Block Bypassing
- Filter Lists

If you think you might have messed up the settings of the System DNS Client, just delete the profile below to reset it to the defaults.
`

	// PortmasterProfileID is the profile ID used for the Portmaster Core itself.
	PortmasterProfileID = "_portmaster"
	// PortmasterProfileName is the name used for the Portmaster Core itself.
	PortmasterProfileName = "Portmaster Core Service"
	// PortmasterProfileDescription is the description used for the Portmaster Core itself.
	PortmasterProfileDescription = `This is the Portmaster itself, which runs in the background as a system service. App specific settings have no effect.`

	// PortmasterAppProfileID is the profile ID used for the Portmaster App.
	PortmasterAppProfileID = "_portmaster-app"
	// PortmasterAppProfileName is the name used for the Portmaster App.
	PortmasterAppProfileName = "Portmaster User Interface"
	// PortmasterAppProfileDescription is the description used for the Portmaster App.
	PortmasterAppProfileDescription = `This is the Portmaster UI Windows.`

	// PortmasterNotifierProfileID is the profile ID used for the Portmaster Notifier.
	PortmasterNotifierProfileID = "_portmaster-notifier"
	// PortmasterNotifierProfileName is the name used for the Portmaster Notifier.
	PortmasterNotifierProfileName = "Portmaster Notifier"
	// PortmasterNotifierProfileDescription is the description used for the Portmaster Notifier.
	PortmasterNotifierProfileDescription = `This is the Portmaster UI Tray Notifier.`
)

func updateSpecialProfileMetadata(profile *Profile, binaryPath string) (ok, changed bool) {
	// Get new profile name and check if profile is applicable to special handling.
	var newProfileName, newDescription string
	switch profile.ID {
	case UnidentifiedProfileID:
		newProfileName = UnidentifiedProfileName
		newDescription = UnidentifiedProfileDescription
	case SystemProfileID:
		newProfileName = SystemProfileName
		newDescription = SystemProfileDescription
	case SystemResolverProfileID:
		newProfileName = SystemResolverProfileName
		newDescription = SystemResolverProfileDescription
	case PortmasterProfileID:
		newProfileName = PortmasterProfileName
		newDescription = PortmasterProfileDescription
	case PortmasterAppProfileID:
		newProfileName = PortmasterAppProfileName
		newDescription = PortmasterAppProfileDescription
	case PortmasterNotifierProfileID:
		newProfileName = PortmasterNotifierProfileName
		newDescription = PortmasterNotifierProfileDescription
	default:
		return false, false
	}

	// Update profile name if needed.
	if profile.Name != newProfileName {
		profile.Name = newProfileName
		changed = true
	}

	// Update description if needed.
	if profile.Description != newDescription {
		profile.Description = newDescription
		changed = true
	}

	// Update LinkedPath to new value.
	if profile.LinkedPath != binaryPath {
		profile.LinkedPath = binaryPath
		changed = true
	}

	return true, changed
}

func getSpecialProfile(profileID, linkedPath string) *Profile {
	switch profileID {
	case UnidentifiedProfileID:
		return New(SourceLocal, UnidentifiedProfileID, linkedPath, nil)

	case SystemProfileID:
		return New(SourceLocal, SystemProfileID, linkedPath, nil)

	case SystemResolverProfileID:
		systemResolverProfile := New(
			SourceLocal,
			SystemResolverProfileID,
			linkedPath,
			map[string]interface{}{
				// Explicitly setting the default action to "permit" will improve the
				// user experience for people who set the global default to "prompt".
				// Resolved domain from the system resolver are checked again when
				// attributed to a connection of a regular process. Otherwise, users
				// would see two connection prompts for the same domain.
				CfgOptionDefaultActionKey: "permit",
				// Explicitly allow localhost and answers to multicast protocols that
				// are commonly used by system resolvers.
				// TODO: When the Portmaster gains the ability to attribute multicast
				// responses to their requests, these rules can probably be removed
				// again.
				CfgOptionServiceEndpointsKey: []string{
					"+ Localhost",    // Allow everything from localhost.
					"+ LAN UDP/5353", // Allow inbound mDNS requests and multicast replies.
					"+ LAN UDP/5355", // Allow inbound LLMNR requests and multicast replies.
					"+ LAN UDP/1900", // Allow inbound SSDP requests and multicast replies.
				},
				// Explicitly disable all filter lists, as these will be checked later
				// with the attributed connection. As this is the system resolver, this
				// list can instead be used as a global enforcement of filter lists, if
				// the system resolver is used. Users who want to
				CfgOptionFilterListsKey: []string{},
			},
		)
		return systemResolverProfile

	case PortmasterProfileID:
		profile := New(SourceLocal, PortmasterProfileID, linkedPath, nil)
		profile.Internal = true
		return profile

	case PortmasterAppProfileID:
		profile := New(
			SourceLocal,
			PortmasterAppProfileID,
			linkedPath,
			map[string]interface{}{
				CfgOptionDefaultActionKey: "block",
				CfgOptionEndpointsKey: []string{
					"+ Localhost",
					"+ .safing.io",
				},
			},
		)
		profile.Internal = true
		return profile

	case PortmasterNotifierProfileID:
		profile := New(
			SourceLocal,
			PortmasterNotifierProfileID,
			linkedPath,
			map[string]interface{}{
				CfgOptionDefaultActionKey: "block",
				CfgOptionEndpointsKey: []string{
					"+ Localhost",
				},
			},
		)
		profile.Internal = true
		return profile

	default:
		return nil
	}
}

// specialProfileNeedsReset is used as a workaround until we can properly use
// profile layering in a way that it is also correctly handled by the UI. We
// check if the special profile has not been changed by the user and if not,
// check if the profile is outdated and can be upgraded.
func specialProfileNeedsReset(profile *Profile) bool {
	if profile == nil {
		return false
	}

	switch {
	case profile.Source != SourceLocal:
		// Special profiles live in the local scope only.
		return false
	case profile.LastEdited > 0:
		// Profile was edited - don't override user settings.
		return false
	}

	switch profile.ID {
	case SystemResolverProfileID:
		return canBeUpgraded(profile, "20.11.2021")
	case PortmasterAppProfileID:
		return canBeUpgraded(profile, "8.9.2021")
	default:
		// Not a special profile or no upgrade available yet.
		return false
	}
}

func canBeUpgraded(profile *Profile, upgradeDate string) bool {
	// Parse upgrade date.
	upgradeTime, err := time.Parse("2.1.2006", upgradeDate)
	if err != nil {
		log.Warningf("profile: failed to parse date %q: %s", upgradeDate, err)
		return false
	}

	// Check if the upgrade is applicable.
	if profile.Created < upgradeTime.Unix() {
		log.Infof("profile: upgrading special profile %s", profile.ScopedID())
		return true
	}

	return false
}
