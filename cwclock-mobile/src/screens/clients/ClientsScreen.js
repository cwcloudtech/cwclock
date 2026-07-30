import React, { useEffect } from "react";
import { FlatList, StyleSheet, Text, TouchableOpacity } from "react-native";
import { useDispatch, useSelector } from "react-redux";
import Screen from "../../components/Screen";
import { useI18n } from "../../i18n/I18nContext";
import theme from "../../theme";
import { listClientsApi } from "../../redux/clients/clients.actions";

const ClientsScreen = ({ navigation }) => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const { orgId } = useSelector((state) => state.session);
  const { items: clients } = useSelector((state) => state.clients);

  useEffect(() => {
    if (orgId) dispatch(listClientsApi(orgId));
  }, [orgId, dispatch]);

  return (
    <Screen scroll={false}>
      <FlatList
        data={clients}
        keyExtractor={(c) => c.id}
        contentContainerStyle={styles.list}
        ListEmptyComponent={<Text style={styles.empty}>{t("clients.noClients")}</Text>}
        renderItem={({ item }) => (
          <TouchableOpacity style={styles.row} onPress={() => navigation.navigate("ClientForm", { client: item })}>
            <Text style={styles.rowText}>{item.name}</Text>
            {item.email ? <Text style={styles.rowSub}>{item.email}</Text> : null}
          </TouchableOpacity>
        )}
      />
      <TouchableOpacity style={styles.fab} onPress={() => navigation.navigate("ClientForm", {})}>
        <Text style={styles.fabText}>+</Text>
      </TouchableOpacity>
    </Screen>
  );
};

const styles = StyleSheet.create({
  list: {
    padding: theme.spacing(2),
    paddingBottom: theme.spacing(10),
  },
  row: {
    paddingVertical: theme.spacing(1.5),
    borderBottomWidth: 1,
    borderBottomColor: theme.color.border,
  },
  rowText: {
    fontSize: 16,
    color: theme.color.text,
  },
  rowSub: {
    fontSize: 13,
    color: theme.color.textMuted,
    marginTop: 2,
  },
  empty: {
    color: theme.color.textMuted,
    paddingVertical: theme.spacing(2),
  },
  fab: {
    position: "absolute",
    right: theme.spacing(2),
    bottom: theme.spacing(2),
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: theme.color.primary,
    alignItems: "center",
    justifyContent: "center",
    elevation: 4,
  },
  fabText: {
    color: theme.color.white,
    fontSize: 28,
    lineHeight: 30,
  },
});

export default ClientsScreen;
